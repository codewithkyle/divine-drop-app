import db from "@codewithkyle/jsql";
import {navigateTo} from "@codewithkyle/router";
import SuperComponent from "@codewithkyle/supercomponent";
import {html, render} from "lit-html";
import env from "~brixi/controllers/env";
import type { Deck } from "types/cards";
import dayjs from "dayjs";
import CardBrowser from "components/card-browser/card-browser";
import CardFilters from "components/card-filters/card-filters";

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
        const deck = await db.query<Deck>("SELECT * FROM decks WHERE id = $id", {
            id: this.model.deckId,
        });
        if (!deck.length){
            navigateTo("/decks");
        }
        this.set({
            deck: deck[0],
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
        const mystics = await db.query("SELECT COUNT(*) FROM cards WHERE id IN $cards AND rarity = mystic", { cards: this.model.deck.cards });
        const commons = await db.query("SELECT COUNT(*) FROM cards WHERE id IN $cards AND rarity = common", { cards: this.model.deck.cards });
        const rares = await db.query("SELECT COUNT(*) FROM cards WHERE id IN $cards AND rarity = rare", { cards: this.model.deck.cards });
        const uncommons = await db.query("SELECT COUNT(*) FROM cards WHERE id IN $cards AND rarity = uncommon", { cards: this.model.deck.cards });
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
                        <span style="color:#ff9f00;" class="font-2xl block font-medium">${mystics[0]["COUNT(*)"]}</span>
                        <span class="font-xs font-grey-600 block">Mythics</span>
                    </div>
                    <div class="inline-block mr-3">
                        <span style="color:#e6a52f;" class="font-2xl block font-medium">${rares[0]["COUNT(*)"]}</span>
                        <span class="font-xs font-grey-600 block">Rares</span>
                    </div>
                    <div class="inline-block mr-3">
                        <span style="color:#9b97ae;" class="font-2xl block font-medium">${uncommons[0]["COUNT(*)"]}</span>
                        <span class="font-xs font-grey-600 block">Uncommons</span>
                    </div>
                    <div class="inline-block">
                        <span style="color:#9270a3;" class="font-2xl block font-medium">${commons[0]["COUNT(*)"]}</span>
                        <span class="font-xs font-grey-600 block">Commons</span>
                    </div>
                </div>
            </header>
        `;
    }

    private async renderDeck(){
        if (!this.model.deck.cards.length){
            return html`
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
        }
    }

    async render(){
        const view = html`
            ${await this.renderHeader()}
            <deck-builder>
                <div>
                    ${new CardFilters()}
                    ${new CardBrowser()}
                </div>
                <deck-component>
                    ${await this.renderDeck()}
                </deck-component>
            </deck-builder>
        `;
        render(view, this);
    }
}
env.bind("edit-deck-page", EditDeckPage);
