import db from "@codewithkyle/jsql";
import {navigateTo} from "@codewithkyle/router";
import SuperComponent from "@codewithkyle/supercomponent";
import {html, render} from "lit-html";
import env from "~brixi/controllers/env";
import type { Deck } from "types/cards";
import dayjs from "dayjs";
import CardBrowser from "components/card-browser/card-browser";
import CardFilters from "components/card-filters/card-filters";
import DeckComponent from "components/deck-component/deck-component";
import { publish, subscribe, unsubscribe } from "@codewithkyle/pubsub";
import editor from "controllers/editor";

interface IEditDeckPage {
    deck: Deck
    deckId: string,
    updatedNow: boolean,
}
export default class EditDeckPage extends SuperComponent<IEditDeckPage>{
    private ticket:string;

    constructor(tokens, params){
        super();
        this.model = {
            deckId: tokens["ID"],
            deck: null,
            updatedNow: false,
        };
        this.ticket = subscribe("deck-editor", this.inbox.bind(this));
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

    disconnected(): void {
        unsubscribe(this.ticket);
    }

    private handleLabelInput = async (e) => {
        const input = e.currentTarget;
        const value = input.value.trim();
        await editor.updateLabel(value, this.model.deckId);
        const updated = this.get();
        updated.deck.label = value;
        updated.updatedNow = true;
        this.set(updated, true);
        publish("deck-editor", {
            type: "sync",
            data: {...updated.deck},
        });
    };

    private async renderHeader(){
        let rareCount = 0;
        let mysticCount = 0;
        let uncommonCount = 0;
        let commonCount = 0;
        let cardCount = 0;
        for (let i = 0; i < this.model.deck.cards.length; i++){
            cardCount += this.model.deck.cards[i].count;
            switch (this.model.deck.cards[i].rarity){
                case "rare":
                    rareCount += this.model.deck.cards[i].count;
                    break;
                case "mystic":
                    mysticCount += this.model.deck.cards[i].count;
                    break;
                case "uncommon":
                    uncommonCount += this.model.deck.cards[i].count;
                    break;
                case "common":
                    commonCount += this.model.deck.cards[i].count;
                    break;
                default:
                    break;
            }
        }
        return html`
            <header>
                <div flex="row wrap items-center">
                    <span class="bg-grey-200 font-grey-700 line-none radius-0.5 font-medium px-0.5 py-0.25 mr-0.5 inline-block">${cardCount}</span>
                    <input type="text" .value=${this.model.deck.label} @blur=${this.handleLabelInput}>
                    <div flex="row nowrap items-center" class="w-full mt-0.5 px-0.25">
                        <span class="font-sm font-grey-500">Created - ${dayjs(this.model.deck.dateCreated).format("MMM D, YYYY")}</span>
                        <span class="font-xs font-grey-500 mx-0.75">|</span>
                        <span class="font-sm font-grey-500">Updated - ${this.model.updatedNow ? "now" : dayjs(this.model.deck.dateUpdated).format("MMM D, YYYY")}</span>
                    </div>
                </div>
                <div class="bg-grey-100 px-3 py-2 radius-0.5">
                    <div class="inline-block mr-3">
                        <span style="color:#ff9f00;" class="font-2xl block font-medium">${mysticCount}</span>
                        <span class="font-xs font-grey-600 block">Mythics</span>
                    </div>
                    <div class="inline-block mr-3">
                        <span style="color:#e6a52f;" class="font-2xl block font-medium">${rareCount}</span>
                        <span class="font-xs font-grey-600 block">Rares</span>
                    </div>
                    <div class="inline-block mr-3">
                        <span style="color:#9b97ae;" class="font-2xl block font-medium">${uncommonCount}</span>
                        <span class="font-xs font-grey-600 block">Uncommons</span>
                    </div>
                    <div class="inline-block">
                        <span style="color:#9270a3;" class="font-2xl block font-medium">${commonCount}</span>
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
