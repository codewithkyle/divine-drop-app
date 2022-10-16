import SuperComponent from "@codewithkyle/supercomponent";
import { html, render } from "lit-html";
import env from "~brixi/controllers/env";
import Input from "~brixi/components/inputs/input/input";
import Select from "~brixi/components/select/select";
import Chips from "~brixi/components/chips/chips";
import { subscribe, unsubscribe } from "@codewithkyle/pubsub";
import editor from "controllers/editor";
import db from "@codewithkyle/jsql";
import {until} from "lit-html/directives/until";
import {unsafeHTML} from "lit-html/directives/unsafe-html";
import {parse} from "utils/symbols";

interface ICardFilters {
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
export default class CardFilters extends SuperComponent<ICardFilters>{
    private ticket:string;
    private chipsEl: Chips;
    private keywordEl: Select;
    private subtypeEl: Select;

    constructor(){
        super();
        this.model = {
            query: "",
            sort: "",
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
        this.chipsEl = null;
        this.keywordEl = null;
        this.subtypeEl = null;
        this.ticket = subscribe("deck-editor", this.inbox.bind(this));
    }
    async connected(){
        await env.css(["card-filters"]);
        this.render();
    }

    disconnected(): void {
        unsubscribe(this.ticket);
    }

    private inbox(data){
        this.set(data, true);
    }

    private handleSearch(value){
        editor.setQuery(value.trim());
    }

    private handleManaColorChange = (e) => {
        const input = e.currentTarget as HTMLInputElement;
        const value = input.value;
        editor.setColor(value, input.checked);
    }

    private handleSort(value:string){
        editor.setSort(value)
    }

    private setType(value:string){
        editor.setType(value);
    }

    private setLegality(mode){
        editor.setLegality(mode);
    }

    private setRarity(value){
        editor.setRarity(value);
    }

    private addSubtype = (e) => {
        const target = e.currentTarget;
        const value = target.dataset.value;
        if (value !== ""){
            editor.addSubtype(value);
            this.chipsEl.addChip({
                label: value,
                name: value,
            });
        }
    }

    private removeChip(value:string){
        if (this.model.keywords.includes(value)){
            editor.removeKeyword(value);
        } else if (this.model.subtypes.includes(value)){
            editor.removeSubtype(value);
        }
    }

    private addKeyword = (e) => {
        const target = e.currentTarget;
        const value = target.dataset.value;
        if (value !== ""){
            editor.addKeyword(value);
            this.chipsEl.addChip({
                label: value,
                name: value,
            });
        }
    }

    async render(){
        const view = html`
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
                ${new Select({
                    name: "sort",
                    icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><line x1="4" y1="6" x2="13" y2="6"></line><line x1="4" y1="12" x2="11" y2="12"></line><line x1="4" y1="18" x2="11" y2="18"></line><polyline points="15 15 18 18 21 15"></polyline><line x1="18" y1="6" x2="18" y2="18"></line></svg>`,
                    options: [
                        { label: "Name", value: "name" },
                        { label: "Mana Cost (low - high)", value: "totalManaCost" },
                        { label: "Mana Cost (high - low)", value: "totalManaCost DESC" },
                        { label: "Power (low - high)", value: "power" },
                        { label: "Power (high - low)", value: "power DESC" },
                        { label: "Toughness (low - high)", value: "toughness" },
                        { label: "Toughness (high - low)", value: "toughness DESC" },
                    ],
                    value: this.model.sort,
                    class: "ml-1",
                    css: "width:auto;",
                    callback: this.handleSort.bind(this),
                })}
            </div>
            <div class="w-full" flex="row nowrap items-center">
                ${until(
                    db.query("SELECT UNIQUE type FROM cards").then(types => {
                        return new Select({
                            name: "type",
                            value: this.model.type,
                            css: "flex:1;",
                            options: [{ label: "All types", value: null}, ...(types.map((type) => {
                                return {
                                    label: type,
                                    value: type,
                                }
                            }))],
                            class: "mr-1",
                            callback: this.setType.bind(this),
                        })
                    }),
                    html`
                        <div class="skeleton -button w-full mr-1"></div>
                    `
                )}
                ${until(
                    db.query<string>("SELECT UNIQUE rarity FROM cards").then(results => {
                        const rarities = [{ label: "All rarities", value: null }];
                        for (const rarity of results){
                            rarities.push({
                                label: rarity,
                                value: rarity,
                            });
                        }
                        return new Select({
                            name: "rarity",
                            options: rarities,
                            callback: this.setRarity.bind(this),
                            class: "mr-1 w-auto",
                            css: "flex:1;",
                            value: this.model.rarity,
                        });
                    }),
                    html`
                        <div class="skeleton -button w-full mr-1"></div>
                    `
                )}
                ${until(
                    db.query("SELECT legalities FROM cards LIMIT 1").then(results => {
                        const modes = [{ label: "No restrictions", value: null }];
                        for (const key in results[0]["legalities"]){
                            modes.push({
                                label: key,
                                value: key,
                            });
                        }
                        return new Select({
                            name: "legality",
                            options: modes,
                            callback: this.setLegality.bind(this),
                            class: "mr-1 w-auto",
                            css: "flex:1;",
                            value: this.model.legality,
                        });
                    }),
                    html`
                        <div class="skeleton -button w-full mr-1"></div>
                    `
                )}
                <div class="mana">
                    <input @change=${this.handleManaColorChange} type="checkbox" value="white" id="white">
                    <label tooltip="White" for="white">${unsafeHTML(parse("{W}"))}</label>
                    <input @change=${this.handleManaColorChange} type="checkbox" value="black" id="black">
                    <label tooltip="Black" for="black">${unsafeHTML(parse("{B}"))}</label>
                    <input @change=${this.handleManaColorChange} type="checkbox" value="blue" id="blue">
                    <label tooltip="Blue" for="blue">${unsafeHTML(parse("{U}"))}</label>
                    <input @change=${this.handleManaColorChange} type="checkbox" value="red" id="red">
                    <label tooltip="Red" for="red">${unsafeHTML(parse("{R}"))}</label>
                    <input @change=${this.handleManaColorChange} type="checkbox" value="green" id="green">
                    <label tooltip="Green" for="green">${unsafeHTML(parse("{G}"))}</label>
                </div>
            </div>
            <div flex="row nowrap items-center">
                ${until(
                    db.query("SELECT UNIQUE subtypes FROM cards").then(subtypes => {
                        return html`
                            <custom-select class="mr-1" tabindex="0" role="button">
                                <span>Filter By Subtypes</span>
                                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                                   <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                                   <polyline points="8 9 12 5 16 9"></polyline>
                                   <polyline points="16 15 12 19 8 15"></polyline>
                                </svg>
                                <custom-select-options>
                                    ${subtypes.map(subtype => {
                                        return html`
                                            <button @click=${this.addSubtype} data-value="${subtype}">${subtype}</button>
                                        `;
                                    })}
                                </custom-select-options>
                            </custom-select>
                        `;
                    }),
                    html`
                        <div class="skeleton -button mr-1" style="width:200px;"></div>
                    `
                )}
                ${until(
                    db.query("SELECT UNIQUE keywords FROM cards").then(keywords => {
                        return html`
                            <custom-select tabindex="0" role="button">
                                <span>Filter By Keywords</span>
                                <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                                   <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                                   <polyline points="8 9 12 5 16 9"></polyline>
                                   <polyline points="16 15 12 19 8 15"></polyline>
                                </svg>
                                <custom-select-options>
                                    ${keywords.map(keyword => {
                                        return html`
                                            <button @click=${this.addKeyword} data-value="${keyword}">${keyword}</button>
                                        `;
                                    })}
                                </custom-select-options>
                            </custom-select>
                        `;
                    }),
                    html`
                        <div class="skeleton -button mr-1" style="width:200px;"></div>
                    `
                )}
            </div>
            ${new Chips({
                callback: this.removeChip.bind(this),
                type: "dynamic",
                css: "flex:1;",
                class: "pt-0.125",
                chips: [...(this.model.subtypes.map(type => {
                    return {
                        label: type,
                        name: type,
                    };
                })), ...(this.model.keywords.map(type => {
                    return {
                        label: type,
                        name: type,
                    };
                }))],
            })}
        `;
        render(view, this);
        setTimeout(()=>{
            this.chipsEl = this.querySelector("chips-component");
        }, 100);
    }
}
env.bind("card-filters", CardFilters);
