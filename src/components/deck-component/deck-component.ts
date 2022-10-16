import SuperComponent from "@codewithkyle/supercomponent";
import {html, render} from "lit-html";
import env from "~brixi/controllers/env";

interface IDeckComponent {}
export default class DeckComponent extends SuperComponent<IDeckComponent>{
    private mainEl: HTMLElement;
    private height: number;

    constructor(){
        super();
        this.mainEl = document.body.querySelector("main");
    }

    async connected(){
        await env.css(["deck-component"]);
        this.mainEl.addEventListener("scroll", this.handleScroll);
        this.render();
    }

    private handleScroll = (e) => {
        // Header: 211px
        const height = window.innerHeight - 64 - 211 + this.mainEl.scrollTop;
        const maxHeight = window.innerHeight - 64;
        this.style.height = `${height <= maxHeight ? height : maxHeight}px`;
    }

    render(){
        const view = html`
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
        render(view, this);
    }
}
env.bind("deck-component", DeckComponent);
