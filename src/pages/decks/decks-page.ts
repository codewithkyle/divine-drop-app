import db from "@codewithkyle/jsql";
import {navigateTo} from "@codewithkyle/router";
import SuperComponent from "@codewithkyle/supercomponent";
import {UUID} from "@codewithkyle/uuid";
import {html, render} from "lit-html";
import env from "~brixi/controllers/env";
import Button from "~brixi/components/buttons/button/button";
import type { Deck, Card } from "types/cards";
import editor from "controllers/editor";

interface IDecksPage{}
export default class DecksPage extends SuperComponent<IDecksPage>{
    constructor(){
        super();
    }
    async connected(){
        await env.css(["decks-page", "skeletons"]);
        this.render();
    }

    private editDeck = (e) => {
        const target = e.currentTarget;
        const id = target.dataset.id;
        navigateTo(`/deck/${id}`);
    }

    async render(){
        const decks = await db.query<Deck>("SELECT * FROM decks");
        const view = html`
            <div class="w-full p-2 bg-grey-100 border-b-solid border-b-1 border-b-grey-300" flex="row nowrap items-center">
                ${new Button({
                    kind: "solid",
                    color: "white",
                    label: "Create Deck",
                    class: "mr-1",
                    icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><rect x="4" y="4" width="6" height="6" rx="1"></rect><rect x="14" y="4" width="6" height="6" rx="1"></rect><rect x="4" y="14" width="6" height="6" rx="1"></rect><path d="M14 17h6m-3 -3v6"></path></svg>`,
                    iconPosition: "left",
                    callback: editor.createDeck
                })}
                ${new Button({
                    kind: "solid",
                    color: "white",
                    label: "Import Deck",
                    icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><ellipse cx="12" cy="6" rx="8" ry="3"></ellipse><path d="M4 6v8m5.009 .783c.924 .14 1.933 .217 2.991 .217c4.418 0 8 -1.343 8 -3v-6"></path><path d="M11.252 20.987c.246 .009 .496 .013 .748 .013c4.418 0 8 -1.343 8 -3v-6m-18 7h7m-3 -3l3 3l-3 3"></path></svg>`,
                    iconPosition: "left",
                    callback: async () => {
                        await editor.importDeck();
                        this.render();
                    }
                })}
            </div>
            <deck-browser>
                ${await Promise.all(decks.map(async (deck, index) => {
                    let cardImage = null;
                    if (deck.commanderId || deck.cards.length){
                        const card = await db.query<Card>("SELECT front FROM cards WHERE id = $id", {
                            id: deck.commanderId || deck.cards?.[0].id
                        });
                        cardImage = card[0].front;
                    }
                    return html`
                        <button @click=${this.editDeck} data-id="${deck.id}" class="deck">
                            <span class="font-grey-800 block font-medium mb-1.5 line-snug">${deck.label}</span>
                            ${cardImage ? html`
                                <img src="${cardImage}">
                            ` : html`<card-shell class="skeleton"></card-shell>`}
                            <div flex="row nowrap items-center justify-end" class="w-full mt-0.25">
                                ${new Button({
                                    callback: async ()=>{
                                        const result = (await db.query<Deck>("SELECT * FROM decks WHERE id = $id", {
                                            id: deck.id,
                                        }))[0];
                                        delete result.id;
                                        delete result.dateCreated;
                                        delete result.dateUpdated;
                                        window.prompt("Deck Code:", JSON.stringify(result));
                                    },
                                    size: "slim",
                                    icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M14 3v4a1 1 0 0 0 1 1h4"></path><path d="M17 21h-10a2 2 0 0 1 -2 -2v-14a2 2 0 0 1 2 -2h7l5 5v11a2 2 0 0 1 -2 2z"></path><path d="M9 15h6"></path><path d="M12.5 17.5l2.5 -2.5l-2.5 -2.5"></path></svg>`,
                                    kind: "text",
                                    color: "grey",
                                    iconPosition: "center",
                                    tooltip: "Export",
                                })}
                                ${new Button({
                                    callback: async ()=>{
                                        await editor.deleteDeck(deck.id, deck.label);
                                        this.render();
                                    },
                                    size: "slim",
                                    icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><path d="M4 7h16"></path><path d="M5 7l1 12a2 2 0 0 0 2 2h8a2 2 0 0 0 2 -2l1 -12"></path><path d="M9 7v-3a1 1 0 0 1 1 -1h4a1 1 0 0 1 1 1v3"></path><path d="M10 12l4 4m0 -4l-4 4"></path></svg>`,
                                    kind: "text",
                                    color: "danger",
                                    iconPosition: "center",
                                    tooltip: "Delete",
                                })}
                            </div>
                        </button>
                    `;
                }))}
            </deck-browser>
        `;
        render(view, this);
    }
}
env.bind("decks-page", DecksPage);
