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
import Spinner from "~brixi/components/progress/spinner/spinner";

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
}
export default class CardFilters extends SuperComponent<ICardFilters>{
    private ticket:string;
    private chipsEl: Chips;

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
        };
        this.chipsEl = null;
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

    private addSubtypeChip(value:string){
        if (value !== null){
            editor.addSubtype(value);
            this.chipsEl.addChip({
                label: value,
                name: value,
            });
        }
    }

    private removeChip(value:string){
        editor.removeSubtype(value);
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
                        });
                    }),
                    html`
                        <div class="skeleton -button w-full mr-1"></div>
                    `
                )}
                <div class="mana">
                    <input @change=${this.handleManaColorChange} type="checkbox" value="white" id="white">
                    <label tooltip="White" for="white">W</label>
                    <input @change=${this.handleManaColorChange} type="checkbox" value="black" id="black">
                    <label tooltip="Black" for="black">B</label>
                    <input @change=${this.handleManaColorChange} type="checkbox" value="blue" id="blue">
                    <label tooltip="Blue" for="blue">U</label>
                    <input @change=${this.handleManaColorChange} type="checkbox" value="red" id="red">
                    <label tooltip="Red" for="red">R</label>
                    <input @change=${this.handleManaColorChange} type="checkbox" value="green" id="green">
                    <label tooltip="Green" for="green">G</label>
                </div>
            </div>
            <div flex="row nowrap items-center">
                ${until(
                    db.query("SELECT UNIQUE subtypes FROM cards").then(subtypes => {
                        return new Select({
                            name: "type",
                            value: null,
                            css: "width:200px;",
                            class: "mr-0.5",
                            options: [{ label: "Filter by subtype", value: null}, ...(subtypes.map((type) => {
                                return {
                                    label: type,
                                    value: type,
                                }
                            }))],
                            callback: this.addSubtypeChip.bind(this),
                        })
                    }),
                    html`
                        <div class="skeleton -button mr-1" style="width:200px;"></div>
                    `
                )}
                ${new Chips({
                    callback: this.removeChip.bind(this),
                    type: "dynamic",
                    kind: "text",
                    css: "flex:1;",
                    class: "pt-0.125",
                    chips: (this.model.subtypes.map(type => {
                        return {
                            label: type,
                            name: type,
                        };
                    })),
                })}
            </div>
        `;
        render(view, this);
        setTimeout(()=>{
            this.chipsEl = this.querySelector("chips-component");
        }, 80);
    }
}
env.bind("card-filters", CardFilters);
