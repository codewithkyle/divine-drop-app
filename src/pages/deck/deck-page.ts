import SuperComponent from "@codewithkyle/supercomponent";
import {html, render} from "lit-html";
import env from "~brixi/controllers/env";
import type { Deck, Card } from "types/cards";
import db from "@codewithkyle/jsql";
import DeckHeader from "components/deck-header/deck-header";
import {until} from "lit-html/directives/until";
import Sortable from "sortablejs";

interface IDeckPage {
    deck: Deck,
}
export default class DeckPage extends SuperComponent<IDeckPage>{
    private deckId: string;

    constructor(tokens){
        super();
        this.deckId = tokens["ID"];
        this.state = "LOADING";
        this.stateMachine = {
            LOADING: {
                DONE: "IDLING",
                LOAD: "LOADING",
            },
            IDLING: {
                LOAD: "LOADING",
            },
        };
        this.model = {
            deck: null,
        };
    }

    override async connected() {
        await env.css(["deck-page", "skeletons"]);
        this.render();
        const deck = (await db.query<Deck>("SELECT * FROM decks WHERE id = $id", { id: this.deckId }))[0];
        this.set({
            deck: deck,
        });
        this.trigger("DONE");
    }

    private sortable(){
        const container = this.querySelector("[sortable]");
        if (container){
            new Sortable(container, {
                animation: 150,
                onUpdate: e => {
                    console.log(e);
                }
            });
        }
    }

    private renderLoading(){
        return html`
            ${new DeckHeader(this.deckId)}
            <div class="deck">
                ${new Array(30).fill(null).map(() => {
                    return html`<card-shell class="skeleton"></card-shell>`;
                })}
            </div>
        `;
    }

    private renderCards(){
        return html`
            ${new DeckHeader(this.deckId)}
            <div class="deck" sortable>
                ${this.model.deck.cards.map(card => {
                    return until(
                        db.query<Card>("SELECT * FROM cards WHERE id = $id", { id: card.id }).then(cards => {
                            const card:Card = cards[0];
                            return html`
                                <div class="card">
                                    <img src="${card.front}">
                                </div>
                            `;
                        }),
                        html`<card-shell class="skeleton"></card-shell>`,
                    )
                })}
            </div>
        `;
    }

    override render(): void {
        let view;
        switch(this.state){
            case "IDLING":
                view = this.renderCards();
                break;
            default:
                view = this.renderLoading();
                break;
        }
        render(view, this);
        setTimeout(this.sortable.bind(this), 100);
    }
}
env.bind("deck-page", DeckPage);
