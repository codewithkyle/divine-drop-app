import db from "@codewithkyle/jsql";
import SuperComponent from "@codewithkyle/supercomponent";
import {html, render} from "lit-html";
import env from "~brixi/controllers/env";
import type { Deck } from "../../types/cards.d.ts";

interface IEditDeckPage {
    deck: Deck
}
export default class EditDeckPage extends SuperComponent<IEditDeckPage>{
    private deckId: string;

    constructor(tokens, params){
        super();
        this.deckId = tokens["ID"];
        this.model = {
            deck: null,
        };
    }
    async connected(){
        await env.css(["edit-deck-page"]);
        const deck = await db.query<Deck>("SELECT * FROM decks WHERE id = $id", {
            id: this.deckId,
        });
        this.set({
            deck: deck[0],
        });
        this.render();
    }
    render(){
        const view = html`Edit deck '${this.model.deck.label}'`;
        render(view, this);
    }
}
env.bind("edit-deck-page", EditDeckPage);
