import db from "@codewithkyle/jsql";
import SuperComponent from "@codewithkyle/supercomponent";
import {html, render} from "lit-html";
import env from "~brixi/controllers/env";
import Input from "~brixi/types/input";
import Select from "~brixi/types/select";
import type { Card } from "types/cards";
import Pagination from "~brixi/types/pagination";

interface ICardBrowser{
    cards: Card[],
    query: string,
    sort: string,
    page: number,
    totalPages: number,
    colors: {
        white: boolean,
        black: boolean,
        blue: boolean,
        red: boolean,
        green: boolean,
    }
}
export default class CardBrowser extends SuperComponent<ICardBrowser>{
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
            query: "",
            sort: "name",
            page: 1,
            totalPages: 1,
            colors: {
                white: false,
                black: false,
                blue: false,
                red: false,
                green: false,
            }
        };
    }

    async connected(){
        await env.css(["card-browser", "skeletons"]);
        this.render();
        this.queryCards();
    }

    private handleSearch(value){
        this.set({
            query: value.trim(),
            page: 1,
        }, true);
        this.queryCards();
    }

    private handleManaColorChange = (e) => {
        const input = e.currentTarget as HTMLInputElement;
        const value = input.value;
        const updated = this.get();
        updated.colors[value] = input.checked;
        updated.page = 1;
        this.set(updated, true);
        this.queryCards();
    }

    private async queryCards(resetPage = false){
        this.trigger("LOAD");

        const data = {};
        let cardQuery = "SELECT id, front, name FROM cards";
        let countQuery = "SELECT COUNT(*) FROM cards"

        if (this.model.query.length || this.model.colors.black || this.model.colors.blue || this.model.colors.green || this.model.colors.red || this.model.colors.white){
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

        const cards = await db.query<Card>(`${cardQuery} ${conditions.join(" AND ")} OFFSET ${(this.model.page - 1) * 30} LIMIT 30 ORDER BY ${this.model.sort}`, data);
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

    private renderFilters(){
        return html`
            <div class="w-full mb-2" grid="rows 1 gap-1">
                <div flex="row nowrap items-center">
                    ${new Input({
                        name: "search",
                        placeholder: "Search cards",
                        class: "w-full",
                        css: "flex:1;",
                        value: this.model.query,
                        icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><circle cx="10" cy="10" r="7"></circle><line x1="21" y1="21" x2="15" y2="15"></line></svg>`,
                        callback: this.debounce(this.handleSearch.bind(this), 300),
                    })}
                    <div class="ml-1 mana" flex="row nowrap items-center">
                        <input @change=${this.handleManaColorChange} type="checkbox" value="white" id="white">
                        <label for="white">W</label>
                        <input @change=${this.handleManaColorChange} type="checkbox" value="black" id="black">
                        <label for="black">B</label>
                        <input @change=${this.handleManaColorChange} type="checkbox" value="blue" id="blue">
                        <label for="blue">U</label>
                        <input @change=${this.handleManaColorChange} type="checkbox" value="red" id="red">
                        <label for="red">R</label>
                        <input @change=${this.handleManaColorChange} type="checkbox" value="green" id="green">
                        <label for="green">G</label>
                    </div>
                    ${new Select({
                        name: "sort",
                        icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><line x1="4" y1="6" x2="13" y2="6"></line><line x1="4" y1="12" x2="11" y2="12"></line><line x1="4" y1="18" x2="11" y2="18"></line><polyline points="15 15 18 18 21 15"></polyline><line x1="18" y1="6" x2="18" y2="18"></line></svg>`,
                        options: [
                            { label: "Name", value: "name" },
                            { label: "Mana Cost (low - high)", value: "totalManaCost" },
                            { label: "Mana Cost (high - low)", value: "totalManaCost DESC" },
                        ],
                        value: this.model.sort,
                        class: "ml-1",
                        css: "width:auto;",
                        callback: (value) => {
                            this.set({
                                sort: value,
                                page: 1,
                            }, true);
                            this.queryCards();
                        },
                    })}
                </div>
            </div>
        `;
    }

    private renderCards(){
        return html`
            ${this.model.cards.map(card => {
                return html`
                    <button class="card">
                        <img src="${card.front}" draggable="false" onload="this.style.opacity = '1';">
                    </button>
                `;
            })}
        `;
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
            ${this.renderFilters()}
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
