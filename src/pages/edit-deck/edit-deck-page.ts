import db from "@codewithkyle/jsql";
import { navigateTo } from "@codewithkyle/router";
import SuperComponent from "@codewithkyle/supercomponent";
import { html, render } from "lit-html";
import env from "~brixi/controllers/env";
import type { Deck } from "types/cards";
import CardBrowser from "components/card-browser/card-browser";
import CardFilters from "components/card-filters/card-filters";
import DeckComponent from "components/deck-component/deck-component";
import DeckHeader from "components/deck-header/deck-header";

interface IEditDeckPage {}
export default class EditDeckPage extends SuperComponent<IEditDeckPage>{
    private ticket:string;
    private deckId: string;

    constructor(tokens, params){
        super();
        this.deckId = tokens["ID"];
    }

    async connected(){
        await env.css(["edit-deck-page"]);
        const deck = (await db.query<Deck>("SELECT * FROM decks WHERE id = $id", {
            id: this.deckId,
        }))?.[0] ?? null;
        if (!deck){
            navigateTo("/decks");
        }
        this.set({
            deck: deck,
        });
        this.render();
    }

    async render(){
        const view = html`
            ${new DeckHeader(this.deckId)}
            <deck-builder>
                <div>
                    ${new CardFilters()}
                    ${new CardBrowser(this.deckId)}
                </div>
                ${new DeckComponent(this.deckId)}
            </deck-builder>
        `;
        render(view, this);
    }
}
env.bind("edit-deck-page", EditDeckPage);
