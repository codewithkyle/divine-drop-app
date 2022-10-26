import SuperComponent from "@codewithkyle/supercomponent";
import {html, render} from "lit-html";
import env from "~brixi/controllers/env";
import Sortable from "sortablejs";
import type { Deck, Card } from "types/cards";
import db from "@codewithkyle/jsql";
import editor from "controllers/editor";
import { subscribe, unsubscribe } from "@codewithkyle/pubsub";
import {until} from "lit-html/directives/until";
import DeckCard from "./card/deck-card";

interface IDeckComponent {
    deck: Deck,
}
export default class DeckComponent extends SuperComponent<IDeckComponent>{
    private mainEl: HTMLElement;
    private height: number;
    private deckId: string;
    private ticket: string;

    constructor(deckId: string){
        super();
        this.deckId = deckId;
        this.mainEl = document.body.querySelector("main");
        this.model = {
            deck: null,
        };
        this.state = "LOADING";
        this.stateMachine = {
            LOADING: {
                DONE: "IDLING",
                LOAD: "LOADING",
            },
            IDLING: {
                LOAD: "LOADING",
                DONE: "IDLING",
            },
        };
    }

    async connected(){
        await env.css(["deck-component"]);
        this.render();
        this.mainEl.addEventListener("scroll", this.handleScroll);
        const deck = (await db.query<Deck>("SELECT * FROM decks WHERE id = $id", { id: this.deckId }))[0];
        this.set({
            deck: deck,
        });
        this.trigger("DONE");
        this.ticket = subscribe("deck-editor", this.inbox.bind(this));
        this.addEventListener("drop", this.handleDrop);
        this.addEventListener("dragover", (e) => {
            e.preventDefault();
            this.classList.add("has-hover");
        });
        this.addEventListener("dragenter", ()=>{
            this.classList.add("has-hover");
        });
        this.addEventListener("dragleave", ()=>{
            this.classList.remove("has-hover");
        });
        this.addEventListener("dragend", ()=>{
            this.classList.remove("has-hover");
        });
    }

    override disconnected(): void {
        unsubscribe(this.ticket);
    }

    private inbox({ type, data }){
        switch(type){
            case "add":
                const cardEl = this.querySelector(`deck-card[data-id="${data.id}"]`) || new DeckCard(data, this.deckId, 1, this.updateCardCount.bind(this), this.model.deck.commanderId, this.updateCommander.bind(this));
                if (cardEl.isConnected){
                    cardEl.add();
                } else {
                    this.appendChild(cardEl);
                }
                break;
            case "sync":
                this.set({
                    deck: data,
                }, true);
                break;
            default:
                break;
        }
    }

    private handleDrop = async (e) => {
        this.classList.remove("has-hover");
        const cardId = e.dataTransfer.getData("cardId");
        const rarity = e.dataTransfer.getData("rarity");
        if (!cardId || !rarity) return;
        editor.addCard(cardId, rarity, this.deckId);
    }

    private handleScroll = (e) => {
        // Header: 211px
        const height = window.innerHeight - 64 - 211 + this.mainEl.scrollTop;
        const maxHeight = window.innerHeight - 64;
        this.style.height = `${height <= maxHeight ? height : maxHeight}px`;
    }

    private sortable(){
        new Sortable(this, {
            sort: true,
            animation: 150,
            onEnd: (e) => {
                if (e.oldIndex === e.newIndex) return;
                const updated = this.get();
                const temp = updated.deck.cards.splice(e.oldIndex, 1)[0];
                updated.deck.cards.splice(e.newIndex, 0, temp);
                for (let i = updated.deck.cards.length - 1; i >= 0; i--){
                    if (updated.deck.cards[i] == null){
                        updated.deck.cards.splice(i, 1);
                    }
                }
                db.query("UPDATE decks SET $deck WHERE id = $id", {
                    id: updated.deck.id,
                    deck: updated.deck,
                });
                this.set(updated, true);
            }
        });
    }

    public updateCardCount(cardId:string, count:number){
        for (let i = 0; i < this.model.deck.cards.length; i++){
            if (this.model.deck.cards[i].id === cardId){
                this.model.deck.cards[i].count = count;
                break;
            }
        }
        editor.updateDeck(this.model.deck);
    }

    public updateCommander(cardId:string){
        const updated = this.get();
        updated.deck.commanderId = cardId;
        this.set(updated, true);
        editor.setCommander(cardId, this.deckId);
        this.querySelectorAll("deck-card").forEach((el:DeckCard) => {
            el.setCommanderId(cardId);
        });
    }

    render(){
        let view;
        if (this.state === "LOADING"){
            view = html`
                ${new Array(10).fill(null).map(()=>{
                    return html`
                        <div class="px-0.5 pt-0.25" style="height:48px;">
                            <div class="skeleton -heading w-full"></div>
                        </div>
                    `;
                })}
            `
        } else {
            if (!this.model.deck.cards.length){
                view = html`
                    <div class="text-center placeholder">
                        <svg class="font-grey-400 block mx-auto" xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                           <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                           <path d="M19 11v-2a2 2 0 0 0 -2 -2h-8a2 2 0 0 0 -2 2v8a2 2 0 0 0 2 2h2"></path>
                           <path d="M13 13l9 3l-4 2l-2 4l-3 -9"></path>
                           <line x1="3" y1="3" x2="3" y2="3.01"></line>
                           <line x1="7" y1="3" x2="7" y2="3.01"></line>
                           <line x1="11" y1="3" x2="11" y2="3.01"></line>
                           <line x1="15" y1="3" x2="15" y2="3.01"></line>
                           <line x1="3" y1="7" x2="3" y2="7.01"></line>
                           <line x1="3" y1="11" x2="3" y2="11.01"></line>
                           <line x1="3" y1="15" x2="3" y2="15.01"></line>
                        </svg>
                        <span class="font-xs font-grey-400 mt-0.25 block">Drop cards here to begin.</span>
                    </div>
                `;
            } else {
                view = html`
                    ${this.model.deck.cards.map(deckCard => {
                        return html`
                            ${until(
                                db.query<Card>("SELECT * FROM cards WHERE id = $id", { id: deckCard.id }).then(cards => {
                                    const card = cards[0];
                                    return new DeckCard(card, this.deckId, deckCard.count, this.updateCardCount.bind(this), this.model.deck.commanderId, this.updateCommander.bind(this));
                                }),
                                html`
                                    <div class="p-0.25" style="height:48px;">
                                        <div class="skeleton w-full h-full"></div>
                                    </div>
                                `
                            )}
                        `;
                    })}
                `;
            }
        }
        render(view, this);
        setTimeout(this.sortable.bind(this), 100);
    }
}
env.bind("deck-component", DeckComponent);
