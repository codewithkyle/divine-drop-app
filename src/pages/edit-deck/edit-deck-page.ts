import db from "@codewithkyle/jsql";
import {navigateTo} from "@codewithkyle/router";
import SuperComponent from "@codewithkyle/supercomponent";
import {html, render} from "lit-html";
import { until } from "lit-html/directives/until";
import env from "~brixi/controllers/env";
import type { Deck } from "types/cards";
import dayjs from "dayjs";
import CardBrowser from "components/card-browser/card-browser";
import CardFilters from "components/card-filters/card-filters";
import Spinner from "~brixi/components/progress/spinner/spinner";
import DeckComponent from "components/deck-component/deck-component";

interface IEditDeckPage {
    deck: Deck
    deckId: string,
    updatedNow: boolean,
}
export default class EditDeckPage extends SuperComponent<IEditDeckPage>{

    constructor(tokens, params){
        super();
        this.model = {
            deckId: tokens["ID"],
            deck: null,
            updatedNow: false,
        };
    }
    async connected(){
        await env.css(["edit-deck-page"]);
        const deck = (await db.query<Deck>("SELECT * FROM decks WHERE id = $id", {
            id: this.model.deckId,
        }))?.[0] ?? null;
        if (!deck){
            navigateTo("/decks");
        }
        this.set({
            deck: deck,
        });
        this.render();
    }

     private async updateLabel(value:string){
        await db.query("UPDATE decks SET label = $value WHERE id = $id", {
            value: value.trim(),
            id: this.model.deck.id,
        });
        this.set({
            updatedNow: true,
        }, true);
    }
    private debounceLabelInput = this.debounce(this.updateLabel.bind(this), 300);
    private handleLabelInput = (e) => {
        const input = e.currentTarget;
        const value = input.value.trim();
        this.debounceLabelInput(value);
    };

    private async renderHeader(){
        return html`
            <header>
                <div flex="row wrap items-center">
                    <span class="bg-grey-200 font-grey-700 line-none radius-0.5 font-medium px-0.5 py-0.25 mr-0.5 inline-block">${this.model.deck.cards.length}</span>
                    <input type="text" .value=${this.model.deck.label} @input=${this.handleLabelInput}>
                    <div flex="row nowrap items-center" class="w-full mt-0.5 px-0.25">
                        <span class="font-sm font-grey-500">Created - ${dayjs(this.model.deck.dateCreated).format("MMM D, YYYY")}</span>
                        <span class="font-xs font-grey-500 mx-0.75">|</span>
                        <span class="font-sm font-grey-500">Updated - ${this.model.updatedNow ? "now" : dayjs(this.model.deck.dateUpdated).format("MMM D, YYYY")}</span>
                    </div>
                </div>
                <div class="bg-grey-100 px-3 py-2 radius-0.5">
                    <div class="inline-block mr-3">
                        <span style="color:#ff9f00;" class="font-2xl block font-medium">${until(
                            db.query("SELECT COUNT(*) FROM cards WHERE id IN $cards AND rarity = mystic", { cards: this.model.deck.cards }).then(results => {
                                return results[0]["COUNT(*)"];
                            }),
                            new Spinner({
                                size: 24,
                                color: "grey",
                            })
                        )}</span>
                        <span class="font-xs font-grey-600 block">Mythics</span>
                    </div>
                    <div class="inline-block mr-3">
                        <span style="color:#e6a52f;" class="font-2xl block font-medium">${until(
                            db.query("SELECT COUNT(*) FROM cards WHERE id IN $cards AND rarity = rare", { cards: this.model.deck.cards }).then(results => {
                                return results[0]["COUNT(*)"];
                            }),
                            new Spinner({
                                size: 24,
                                color: "grey",
                            })
                        )}</span>
                        <span class="font-xs font-grey-600 block">Rares</span>
                    </div>
                    <div class="inline-block mr-3">
                        <span style="color:#9b97ae;" class="font-2xl block font-medium">${until(
                            db.query("SELECT COUNT(*) FROM cards WHERE id IN $cards AND rarity = uncommon", { cards: this.model.deck.cards }).then(results => {
                                return results[0]["COUNT(*)"];
                            }), 
                            new Spinner({
                                size: 24,
                                color: "grey",
                            })
                        )}</span>
                        <span class="font-xs font-grey-600 block">Uncommons</span>
                    </div>
                    <div class="inline-block">
                        <span style="color:#9270a3;" class="font-2xl block font-medium">${until(
                            db.query("SELECT COUNT(*) FROM cards WHERE id IN $cards AND rarity = common", { cards: this.model.deck.cards }).then(results => {
                                return results[0]["COUNT(*)"];
                            }),
                            new Spinner({
                                size: 24,
                                color: "grey",
                            })
                        )}</span>
                        <span class="font-xs font-grey-600 block">Commons</span>
                    </div>
                </div>
            </header>
        `;
    }

    async render(){
        const view = html`
            ${await this.renderHeader()}
            <deck-builder>
                <div>
                    ${new CardFilters()}
                    ${new CardBrowser()}
                </div>
                ${new DeckComponent()}
            </deck-builder>
        `;
        render(view, this);
    }
}
env.bind("edit-deck-page", EditDeckPage);
