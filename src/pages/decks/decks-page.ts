import db from "@codewithkyle/jsql";
import {navigateTo} from "@codewithkyle/router";
import SuperComponent from "@codewithkyle/supercomponent";
import {UUID} from "@codewithkyle/uuid";
import {html, render} from "lit-html";
import env from "~brixi/controllers/env";
import Button from "~brixi/components/buttons/button/button";
import type { Deck, Card } from "types/cards";

interface IDecksPage{}
export default class DecksPage extends SuperComponent<IDecksPage>{
    constructor(){
        super();
    }
    async connected(){
        await env.css(["decks-page", "skeletons"]);
        this.render();
    }

    private createDeck = async (e) => {
        const id = UUID();
        const timestamp = new Date().getTime();
        const deck:Deck = {
            id: id,
            label: "Untitled",
            commanderId: null,
            cards: [],
            dateCreated: timestamp.toString(),
            dateUpdated: timestamp.toString(),
        };
        await db.query("INSERT INTO decks VALUES ($deck)", {
            deck: deck,
        });
        navigateTo(`/deck/${id}`);
    }

    private editDeck = (e) => {
        const target = e.currentTarget;
        const id = target.dataset.id;
        navigateTo(`/deck/${id}`);
    }

    async render(){
        const decks = await db.query<Deck>("SELECT * FROM decks");
        const view = html`
            <button class="create-deck" @click=${this.createDeck}>
                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                   <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                   <rect x="4" y="4" width="16" height="16" rx="2"></rect>
                   <line x1="9" y1="12" x2="15" y2="12"></line>
                   <line x1="12" y1="9" x2="12" y2="15"></line>
                </svg>
                <span>Create Deck</span>
            </button>
            ${await Promise.all(decks.map(async (deck, index) => {
                let cardImage = null;
                if (deck.commanderId || deck.cards.length){
                    const card = await db.query<Card>("SELECT front FROM cards WHERE id = $id", {
                        id: deck.commanderId || deck.cards?.[0]
                    });
                    cardImage = card[0].front;
                }
                return html`
                    <button @click=${this.editDeck} data-id="${deck.id}" class="deck">
                        ${cardImage ? html`
                            <img src="${cardImage}">
                        ` : html`<card-shell class="skeleton"></card-shell>`}
                        <div flex="row nowrap items-center justify-between" class="mt-0.5 w-full pl-0.5">
                            <span class="font-grey-700">${deck.label}</span>
                            <div>
                                ${new Button({
                                    callback: async ()=>{
                                        const input = window.prompt("New Name:");
                                        if (input?.length){
                                            await db.query("UPDATE decks SET label = $name WHERE id = $id", {
                                                name: input,
                                                id: deck.id,
                                            });
                                            this.render();
                                        }
                                    },
                                    size: "slim",
                                    icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M4 20h4l10.5 -10.5a1.5 1.5 0 0 0 -4 -4l-10.5 10.5v4"></path><line x1="13.5" y1="6.5" x2="17.5" y2="10.5"></line></svg>`,
                                    kind: "text",
                                    color: "grey",
                                    iconPosition: "center",
                                    tooltip: "Rename",
                                })}
                                ${new Button({
                                    callback: async ()=>{
                                        const confirmed = window.confirm(`Are you sure you want to delete '${deck.label}'?`);
                                        if (confirmed){
                                            await db.query("DELETE FROM decks WHERE id = $id", {
                                                id: deck.id,
                                            });
                                            this.render();
                                        }
                                    },
                                    size: "slim",
                                    icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M4 7h16"></path><path d="M5 7l1 12a2 2 0 0 0 2 2h8a2 2 0 0 0 2 -2l1 -12"></path><path d="M9 7v-3a1 1 0 0 1 1 -1h4a1 1 0 0 1 1 1v3"></path><path d="M10 12l4 4m0 -4l-4 4"></path></svg>`,
                                    kind: "text",
                                    color: "danger",
                                    iconPosition: "center",
                                    tooltip: "Delete",
                                })}
                            </div>
                        </div>
                    </button>
                `;
            }))}
        `;
        render(view, this);
    }
}
env.bind("decks-page", DecksPage);
