import SuperComponent from "@codewithkyle/supercomponent";
import {html, render} from "lit-html";
import env from "~brixi/controllers/env";
import Sortable from "sortablejs";
import type { Deck, Card } from "types/cards";
import db from "@codewithkyle/jsql";
import editor from "controllers/editor";
import { subscribe, unsubscribe } from "@codewithkyle/pubsub";
import {until} from "lit-html/directives/until";
import Button from "~brixi/components/buttons/button/button";
import CardModal from "components/card-modal/card-modal";

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
        this.ticket = subscribe("deck-editor", this.inbox.bind(this));
    }

    async connected(){
        await env.css(["deck-component"]);
        this.mainEl.addEventListener("scroll", this.handleScroll);
        const deck = (await db.query<Deck>("SELECT * FROM decks WHERE id = $id", { id: this.deckId }))[0];
        this.set({
            deck: deck,
        });
        this.addEventListener("drop", this.handleDrop);
        this.addEventListener("dragover", (e) => e.preventDefault());
    }

    override disconnected(): void {
        unsubscribe(this.ticket);
    }

    private inbox({ type, data }){
        switch(type){
            case "sync":
                this.set({
                    deck: data,
                });
                break;
            default:
                break;
        }
    }

    private handleDrop = async (e) => {
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
            onEnd: (e) => {
                if (e.oldIndex === e.newIndex) return;
                const updated = this.get();
                const temp = updated.deck.cards.splice(e.oldIndex, 1)[0];
                updated.deck.cards.splice(e.newIndex, 0, temp);
                db.query("UPDATE decks SET $deck WHERE id = $id", {
                    id: updated.deck.id,
                    deck: updated.deck,
                });
                this.set(updated, true);
            }
        });
    }

    render(){
        let view;
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
                                return html`
                                    <div class="card">
                                        <img src="${card.front}">
                                        <div flex="row nowrap items-center">
                                            ${new Button({
                                                iconPosition: "center",
                                                kind: "text",
                                                color: "white",
                                                class: "mr-0.5",
                                                tooltip: "Preview",
                                                callback: ()=>{
                                                    const modal = new CardModal(card.id);
                                                    document.body.appendChild(modal);
                                                },
                                                icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><circle cx="12" cy="12" r="2"></circle><path d="M22 12c-2.667 4.667 -6 7 -10 7s-7.333 -2.333 -10 -7c2.667 -4.667 6 -7 10 -7s7.333 2.333 10 7"></path></svg>`,
                                            })}
                                            <span class="font-white font-medium font-sm inline-block" style="flex:1;width: 215px;overflow:hidden;">${card.name}</span>
                                        </div>
                                        <div flex="row nowrap items-center" class="actions">
                                            ${new Button({
                                                iconPosition: "center",
                                                kind: "text",
                                                color: "warning",
                                                callback: ()=>{
                                                    const updated = this.get();
                                                    if (updated.deck.commanderId === card.id){
                                                        updated.deck.commanderId = null;
                                                    } else {
                                                        updated.deck.commanderId = card.id;
                                                    }
                                                    this.set(updated, true);
                                                    editor.updateDeck(updated.deck);
                                                },
                                                tooltip: card.id !== this.model.deck.commanderId ? "Make commander" : "Unset commander",
                                                icon: card.id !== this.model.deck.commanderId ?
                                                        `<svg fill="#FDE047" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 576 512"><path d="M287.9 0C297.1 0 305.5 5.25 309.5 13.52L378.1 154.8L531.4 177.5C540.4 178.8 547.8 185.1 550.7 193.7C553.5 202.4 551.2 211.9 544.8 218.2L433.6 328.4L459.9 483.9C461.4 492.9 457.7 502.1 450.2 507.4C442.8 512.7 432.1 513.4 424.9 509.1L287.9 435.9L150.1 509.1C142.9 513.4 133.1 512.7 125.6 507.4C118.2 502.1 114.5 492.9 115.1 483.9L142.2 328.4L31.11 218.2C24.65 211.9 22.36 202.4 25.2 193.7C28.03 185.1 35.5 178.8 44.49 177.5L197.7 154.8L266.3 13.52C270.4 5.249 278.7 0 287.9 0L287.9 0zM287.9 78.95L235.4 187.2C231.9 194.3 225.1 199.3 217.3 200.5L98.98 217.9L184.9 303C190.4 308.5 192.9 316.4 191.6 324.1L171.4 443.7L276.6 387.5C283.7 383.7 292.2 383.7 299.2 387.5L404.4 443.7L384.2 324.1C382.9 316.4 385.5 308.5 391 303L476.9 217.9L358.6 200.5C350.7 199.3 343.9 194.3 340.5 187.2L287.9 78.95z"/></svg>` :
                                                        `<svg fill="#FDE047" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 576 512"><path d="M316.9 18C311.6 7 300.4 0 288.1 0s-23.4 7-28.8 18L195 150.3 51.4 171.5c-12 1.8-22 10.2-25.7 21.7s-.7 24.2 7.9 32.7L137.8 329 113.2 474.7c-2 12 3 24.2 12.9 31.3s23 8 33.8 2.3l128.3-68.5 128.3 68.5c10.8 5.7 23.9 4.9 33.8-2.3s14.9-19.3 12.9-31.3L438.5 329 542.7 225.9c8.6-8.5 11.7-21.2 7.9-32.7s-13.7-19.9-25.7-21.7L381.2 150.3 316.9 18z"/></svg>`,
                                            })}
                                            ${new Button({
                                                iconPosition: "center",
                                                kind: "text",
                                                color: "danger",
                                                callback: ()=>{
                                                    editor.removeCard(card.id, this.deckId);
                                                },
                                                icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><line x1="4" y1="7" x2="20" y2="7"></line><line x1="10" y1="11" x2="10" y2="17"></line><line x1="14" y1="11" x2="14" y2="17"></line><path d="M5 7l1 12a2 2 0 0 0 2 2h8a2 2 0 0 0 2 -2l1 -12"></path><path d="M9 7v-3a1 1 0 0 1 1 -1h4a1 1 0 0 1 1 1v3"></path></svg>`,
                                            })}
                                            ${new Button({
                                                iconPosition: "center",
                                                kind: "text",
                                                color: "grey",
                                                callback: ()=>{
                                                    const updated = this.get();
                                                    for (let i = 0; i < updated.deck.cards.length; i++){
                                                        if (updated.deck.cards[i].id === card.id){
                                                            updated.deck.cards[i].count -= 1;
                                                            if (updated.deck.cards[i].count < 1){
                                                                updated.deck.cards[i].count = 0;
                                                            }
                                                            break;
                                                        }
                                                    }
                                                    this.set(updated, true);
                                                    editor.updateDeck(updated.deck);
                                                },
                                                icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><line x1="5" y1="12" x2="19" y2="12"></line></svg>`,
                                            })}
                                            ${new Button({
                                                iconPosition: "center",
                                                kind: "text",
                                                color: "grey",
                                                class: "mr-0.5",
                                                callback: ()=>{
                                                    const updated = this.get();
                                                    for (let i = 0; i < updated.deck.cards.length; i++){
                                                        if (updated.deck.cards[i].id === card.id){
                                                            updated.deck.cards[i].count++;
                                                            break;
                                                        }
                                                    }
                                                    this.set(updated, true);
                                                    editor.updateDeck(updated.deck);
                                                },
                                                icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><line x1="12" y1="5" x2="12" y2="19"></line><line x1="5" y1="12" x2="19" y2="12"></line></svg>`,
                                            })}
                                            <span class="font-white font-medium">${deckCard.count}</span>
                                        </div>
                                    </div>
                                `;
                            }),
                            html`
                                <div class="card">
                                    <div class="skeleton -heading w-full"></div>
                                </div>
                            `
                        )}
                    `;
                })}
            `;
        }
        render(view, this);
        setTimeout(this.sortable.bind(this), 100);
    }
}
env.bind("deck-component", DeckComponent);
