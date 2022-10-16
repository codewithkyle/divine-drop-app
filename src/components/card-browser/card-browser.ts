import db from "@codewithkyle/jsql";
import SuperComponent from "@codewithkyle/supercomponent";
import {html, render} from "lit-html";
import env from "~brixi/controllers/env";
import type { Card } from "types/cards";
import Pagination from "~brixi/types/pagination";
import { subscribe, unsubscribe } from "@codewithkyle/pubsub";

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
}
export default class CardBrowser extends SuperComponent<ICardBrowser>{
    private ticket:string;

    constructor(){
        super();
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
        };
        this.ticket = subscribe("deck-editor", this.inbox.bind(this));
    }

    async connected(){
        await env.css(["card-browser", "skeletons"]);
        console.log("conencted");
        this.render();
        this.queryCards();
    }

    disconnected(): void {
        console.log("disconnet");
        unsubscribe(this.ticket);
    }

    private inbox(data){
        console.log(data);
        this.set(data, true);
        this.queryCards(true);
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

        if (this.model.query?.length || this.model.colors.black || this.model.colors.blue || this.model.colors.green || this.model.colors.red || this.model.colors.white || this.model.type?.length || this.model.subtypes.length || this.model.legality){
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

        const cards = await db.query<Card>(`${cardQuery} ${conditions.join(" AND ")} OFFSET ${(this.model.page - 1) * 30} LIMIT 30 ORDER BY ${this.model.sort}`, data, true);
        const cardCount = (await db.query<Card>(`${countQuery} ${conditions.join(" AND ")}`, data))[0]["COUNT(*)"];
        this.set({
            cards: cards,
            totalPages: Math.ceil(cardCount / 30),
        });
        this.trigger("DONE");
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
                        <button class="card">
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
                                top: 0,
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
