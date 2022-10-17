import SuperComponent from "@codewithkyle/supercomponent";
import {html, render} from "lit-html";
import env from "~brixi/controllers/env";
import type { Deck } from "types/cards";
import editor from "controllers/editor";
import { publish, subscribe, unsubscribe } from "@codewithkyle/pubsub";
import dayjs from "dayjs";
import db from "@codewithkyle/jsql";

interface IDeckHeader {
    deck: Deck,
    updatedNow: boolean,
}
export default class DeckHeader extends SuperComponent<IDeckHeader>{
    private deckId: string;
    private ticket: string;

    constructor(deckId: string){
        super();
        this.deckId = deckId;
        this.model = {
            deck: null,
            updatedNow: false,
        };
    }

    override async connected(){
        await env.css(["deck-header"]);
        const deck = (await db.query<Deck>("SELECT * FROM decks WHERE id = $id", { id: this.deckId }))[0];
        this.set({
            deck: deck,
        });
        this.ticket = subscribe("deck-editor", this.inbox.bind(this));
    }

    override disconnected(): void {
        unsubscribe(this.ticket);
    }

    private inbox({ type, data }){
        switch (type){
            case "sync":
                this.set({
                    updatedNow: true,
                    deck: data,
                });
                break;
            default:
                break;
        }
    }

    private handleLabelInput = async (e) => {
        const input = e.currentTarget;
        const value = input.value.trim();
        await editor.updateLabel(value, this.model.deck.id);
        const updated = this.get();
        updated.deck.label = value;
        updated.updatedNow = true;
        this.set(updated, true);
        publish("deck-editor", {
            type: "sync",
            data: {...updated.deck},
        });
    };

    override render(){
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
        const view = html`
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
        `;
        render(view, this);
    }
}
env.bind("deck-header", DeckHeader);
