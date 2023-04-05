import db from "@codewithkyle/jsql";
import SuperComponent from "@codewithkyle/supercomponent";
import editor from "controllers/editor";
import {html, render} from "lit-html";
import Button from "~brixi/components/buttons/button/button";
import env from "~brixi/controllers/env";
import type { Deck } from "types/cards";
import { subscribe, unsubscribe } from "@codewithkyle/pubsub";

interface ISidebarComponent{
    decksOpen: boolean,
}
export default class SidebarComponent extends SuperComponent<ISidebarComponent>{
    private ticket:string;

    constructor(){
        super();
        this.model = {
            decksOpen: false,
        };
    }

    private inbox({ type, data }){
        switch (type){
            case "sync":
                this.render();
                break;
            case "create":
                this.render();
                break;
            case "delete":
                this.render();
                break;
            default:
                break;
        }
    }
    
    async connected(){
        await env.css(["sidebar-component"]);
        const results = await db.query("SELECT COUNT(id) FROM decks");
        if (results[0]["COUNT(id)"] > 0){
            this.model.decksOpen = true;
        }
        this.render();
        this.ticket = subscribe("deck-editor", this.inbox.bind(this));
    }

    disconnected(): void {
        unsubscribe(this.ticket);
    }

    private toggleDeck = (e) => {
        const target = e.currentTarget;
        this.set({
            decksOpen: target.checked,
        });
    }

    private renderDecks(decks:Deck[]){
        return html`
            <input @change=${this.toggleDeck} type="checkbox" ?checked=${this.model.decksOpen} id="decks">
            <label for="decks">
                <div flex="row nowrap items-center">
                    <i class="mr-0.5">
                        <svg viewBox="0 0 24 24" xmlns="http://www.w3.org/2000/svg" fill="none" stroke="currentColor">
                            <polyline stroke-width="2" stroke-linecap="round" stroke-linejoin="round" points="14.3,15.6 14.3,19.9 3.4,19.9 3.4,2.1 14.3,2.1 14.3,10.6 "/>
                            <polyline stroke-width="2" stroke-linecap="round" stroke-linejoin="round" points="14.3,4.1 20.6,4.1 20.6,21.9 9.7,21.9 9.7,19.9 "/>
                            <polygon stroke-width="2" stroke-linecap="round" stroke-linejoin="round" points="8.8,7.5 6.5,11 8.8,14.5 11.2,11 "/>
                            <polygon stroke-width="2" stroke-linecap="round" stroke-linejoin="round" points="17.3,13.1 15,16.6 14.3,15.6 14.3,10.6 15,9.6 "/>
                            <line stroke-width="2" stroke-linecap="round" stroke-linejoin="round" x1="5.3" y1="4.1" x2="5.3" y2="5.3"/>
                            <line stroke-width="2" stroke-linecap="round" stroke-linejoin="round" x1="12.2" y1="16.6" x2="12.2" y2="17.8"/>
                            <line stroke-width="2" stroke-linecap="round" stroke-linejoin="round" x1="18.5" y1="18.7" x2="18.5" y2="19.9"/>
                        </svg>
                    </i>
                    <span>Decks</span>
                </div>
                <i>
                    <svg fill="currentColor" xmlns="http://www.w3.org/2000/svg" viewBox="0 0 256 512"><path d="M246.6 278.6c12.5-12.5 12.5-32.8 0-45.3l-128-128c-9.2-9.2-22.9-11.9-34.9-6.9s-19.8 16.6-19.8 29.6l0 256c0 12.9 7.8 24.6 19.8 29.6s25.7 2.2 34.9-6.9l128-128z"/></svg>
                </i>
            </label>
            ${this.model.decksOpen ? html`
                <a href="/decks" class="deck">
                    <i class="mr-0.5">
                        <svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round">
                           <path stroke="none" d="M0 0h24v24H0z" fill="none"></path>
                           <rect x="4" y="4" width="6" height="6" rx="1"></rect>
                           <rect x="14" y="4" width="6" height="6" rx="1"></rect>
                           <rect x="4" y="14" width="6" height="6" rx="1"></rect>
                           <rect x="14" y="14" width="6" height="6" rx="1"></rect>
                        </svg>
                    </i>
                    <span>Browse Decks</span>
                </a>
            ` : ""}
            ${this.model.decksOpen ? decks.map((deck, index) => {
                let cardCount:number = 0;
                for (let i = 0; i < deck.cards.length; i++){
                    cardCount += deck.cards[i].count;
                }
                return html`
                    <a href="/deck/${deck.id}" class="deck">
                        <count>${cardCount}</count>
                        <span>${deck.label}</span>
                    </a>
                `;
            }) : ""}
            ${this.model.decksOpen ? html`
                <div class="w-full px-1 pb-1 pt-0.5 bg-neutral-700">
                    ${new Button({
                        kind: "text",
                        // @ts-ignore
                        color: "white",
                        label: "Create Deck",
                        class: "w-full pr-1.75",
                        icon: `<svg xmlns="http://www.w3.org/2000/svg" width="24" height="24" viewBox="0 0 24 24" stroke-width="2" stroke="currentColor" fill="none" stroke-linecap="round" stroke-linejoin="round"><path stroke="none" d="M0 0h24v24H0z" fill="none"></path><rect x="4" y="4" width="6" height="6" rx="1"></rect><rect x="14" y="4" width="6" height="6" rx="1"></rect><rect x="4" y="14" width="6" height="6" rx="1"></rect><path d="M14 17h6m-3 -3v6"></path></svg>`,
                        iconPosition: "left",
                        callback: editor.createDeck
                    })}
                </div>
            ` : ""}
        `;
    }

    async render(){
        const decks = await db.query<Deck>("SELECT * FROM decks");
        const view = html`
            <a href="/" id="logo">
                <svg xmlns="http://www.w3.org/2000/svg" width="458.84" height="442.65" viewBox="0 0 458.84 442.65">
                    <defs>
                        <style>
                            .a{fill:none;stroke:#4338ca;stroke-linecap:round;stroke-linejoin:round;stroke-width:24.87px;}
                            .b{fill:#4338ca;}
                        </style>
                    </defs>
                    <path class="a" d="M276.43,159.65S317.67,47.11,385.88,47.11,473,151.17,473,151.17H417.1v82.48s-5.06,68-49,97.5c-48.17,32.38-189,67.9-267.2,79.08" transform="translate(-26.58 -34.67)"/><polyline class="a" points="172.38 360.07 136.54 430.22 39.03 430.22 150.41 257.94"/><polyline class="a" points="265.65 200.13 244.45 234.82 145.79 257.94 81.04 239.06 12.44 150.41 105.7 136 165.83 105.7 274.13 135.77"/><polyline class="a" points="135.38 115.72 99.92 75.26 96.07 29.78 229.03 54.83 263.34 90.29"/><circle class="b" cx="359.3" cy="63.31" r="15.42"/>
                </svg>
                <span>Divine Drop</span>
            </a>
            ${this.renderDecks(decks)}
        `;
        render(view, this);
    }
}
env.bind("sidebar-component", SidebarComponent);
