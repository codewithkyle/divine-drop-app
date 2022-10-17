import db from "@codewithkyle/jsql";
import SuperComponent from "@codewithkyle/supercomponent";
import {html, render} from "lit-html";
import env from "~brixi/controllers/env";
import type { Card } from "types/cards";
import Pagination from "~brixi/types/pagination";
import { subscribe, unsubscribe } from "@codewithkyle/pubsub";
import editor from "controllers/editor";

interface ICardBrowser{
    cards: Card[],
    page: number,
    totalPages: number,
    query: string,
    sort: string,
    colors: {
        white: boolean,
        black: boolean,
        blue: boolean,
        red: boolean,
        green: boolean,
    },
    type: string,
    subtypes: string[],
    legality: string,
    rarity: string,
    keywords: string[],
}
export default class CardBrowser extends SuperComponent<ICardBrowser>{
    private ticket:string;
    private deckId: string;

    constructor(deckId:string){
        super();
        this.deckId = deckId;
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
            cards: [],
            page: 1,
            totalPages: 1,
            query: "",
            sort: "name",
            colors: {
                white: false,
                black: false,
                blue: false,
                red: false,
                green: false,
            },
            type: null,
            subtypes: [],
            legality: null,
            rarity: null,
            keywords: [],
        };
        this.ticket = subscribe("deck-editor", this.inbox.bind(this));
    }

    async connected(){
        await env.css(["card-browser", "skeletons"]);
        this.render();
        this.queryCards();
    }

    disconnected(): void {
        unsubscribe(this.ticket);
    }

    private inbox({ type, data }){
        switch (type){
            case "filter":
                this.set(data, true);
                this.queryCards(true);
                break;
            default:
                break;
        }
    }

    private async queryCards(resetPage = false){
        this.trigger("LOAD");

        if (resetPage){
            this.set({
                page: 1,
            }, true);
        }

        const data = {};
        let cardQuery = "SELECT * FROM cards";
        let countQuery = "SELECT COUNT(*) FROM cards"

        if (this.model.keywords.length || this.model.rarity || this.model.query?.length || this.model.colors.black || this.model.colors.blue || this.model.colors.green || this.model.colors.red || this.model.colors.white || this.model.type?.length || this.model.subtypes.length || this.model.legality){
            cardQuery += " WHERE ";
            countQuery += " WHERE ";
        }

        const conditions = [];

        if (this.model.query?.length){
            data["query"] = this.model.query;
            conditions.push("name LIKE $query");
        }

        if (this.model.colors.black){
            conditions.push("colors INCLUDES B");
        }
        if (this.model.colors.blue){
            conditions.push("colors INCLUDES U");
        }
        if (this.model.colors.red){
            conditions.push("colors INCLUDES R");
        }
        if (this.model.colors.green){
            conditions.push("colors INCLUDES G");
        }
        if (this.model.colors.white){
            conditions.push("colors INCLUDES W");
        }

        if (this.model.type?.length){
            conditions.push("type = $type");
            data["type"] = this.model.type;
        }

        if (this.model.legality !== null){
            conditions.push(`legalities.${this.model.legality} = legal`);
        }

        if (this.model.subtypes.length){
            const typesConditions = [];
            for (let i = 0; i < this.model.subtypes.length; i++){
                typesConditions.push(`subtypes INCLUDES $subtype${i}`);
                data[`subtype${i}`] = this.model.subtypes[i];
            }
            conditions.push(typesConditions.join(" OR "));
        }

        if (this.model.keywords.length){
            for (let i = 0; i < this.model.keywords.length; i++){
                conditions.push(`keywords INCLUDES $keyword${i}`);
                data[`keyword${i}`] = this.model.keywords[i];
            }
        }

        if (this.model.rarity){
            conditions.push("rarity = $rarity");
            data["rarity"] = this.model.rarity;
        }

        const cards = await db.query<Card>(`${cardQuery} ${conditions.join(" AND ")} OFFSET ${(this.model.page - 1) * 30} LIMIT 30 ORDER BY ${this.model.sort}`, data);
        const cardCount = (await db.query<Card>(`${countQuery} ${conditions.join(" AND ")}`, data))[0]["COUNT(*)"];
        this.set({
            cards: cards,
            totalPages: Math.ceil(cardCount / 30),
        });
        this.trigger("DONE");
    }

    private handleDragStart = (e) => {
        e.dataTransfer.setData("cardId", e.currentTarget.dataset.id);
        e.dataTransfer.setData("rarity", e.currentTarget.dataset.rarity);
    }

    private addCardToDeck = (e) => {
        const target = e.currentTarget;
        const cardId = target.dataset.id;
        const rarity = target.dataset.rarity;
        editor.addCard(cardId, rarity, this.deckId);
    }

    private renderLoading(){
        return html`
            ${Array(30).fill(null).map(() => {
                return html`
                    <card-shell class="skeleton"></card-shell>
                `;
            })}
        `;
    }

    private renderCards(){
        if (this.model.cards.length){
            return html`
                ${this.model.cards.map(card => {
                    return html`
                        <button @click=${this.addCardToDeck} draggable="true" class="card" data-id="${card.id}" data-rarity="${card.rarity}" @dragstart=${this.handleDragStart}>
                            <img src="${card.front}" draggable="false" onload="this.style.opacity = '1';">
                        </button>
                    `;
                })}
            `;
        } else {
            return html`
                <div class="block mx-auto text-center absolute center pt-4 mt-4">
                    <svg class="font-grey-400 mb-0.5" xmlns="http://www.w3.org/2000/svg" width="64" height="64" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                       <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                       <path d="M14 3v4a1 1 0 0 0 1 1h4"></path>
                       <path d="M17 21h-10a2 2 0 0 1 -2 -2v-14a2 2 0 0 1 2 -2h7l5 5v11a2 2 0 0 1 -2 2z"></path>
                       <path d="M12 17v.01"></path>
                       <path d="M12 14a1.5 1.5 0 1 0 -1.14 -2.474"></path>
                    </svg>
                    <span class="font-grey-800 font-medium font-xl block mx-auto">No cards found.</span>
                    <span class="block mx-auto font-sm font-grey-700 mt-0.25">Your search didn't match any cards.</span>
                </div>
            `;
        }
    }

    async render(){
        let cards;
        switch(this.state){
            case "IDLING":
                cards = this.renderCards();
                break;
            default:
                cards = this.renderLoading();
                break;
        }
        const view = html`
            <card-grid>
                ${cards}
            </card-grid>
            <div class="w-full mt-2 text-center">
                ${this.model.totalPages > 1 ? html`
                    ${new Pagination({
                        totalPages: this.model.totalPages,
                        activePage: this.model.page,
                        callback: (newPage:number)=>{
                            this.set({
                                page: newPage,
                            }, true);
                            this.queryCards();
                            document.body.querySelector("main").scrollTo({
                                top: 211,
                                left: 0,
                            });
                        }
                    })}
                ` : ""}
            </div>
        `;
        render(view, this);
    }
}
env.bind("card-browser", CardBrowser);
