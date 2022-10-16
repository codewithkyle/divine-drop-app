import SuperComponent from "@codewithkyle/supercomponent";
import { html, render } from "lit-html";
import env from "~brixi/controllers/env";
import Input from "~brixi/components/inputs/input/input";
import Select from "~brixi/components/select/select";
import { subscribe, unsubscribe } from "@codewithkyle/pubsub";
import editor from "controllers/editor";

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
    types: string[],
    subtypes: string[],
}
export default class CardFilters extends SuperComponent<ICardFilters>{
    private ticket:string;

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
            types: [],
            subtypes: [],
        };
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

    render(){
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
        `;
        render(view, this);
    }
}
env.bind("card-filters", CardFilters);
