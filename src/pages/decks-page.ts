import SuperComponent from "@codewithkyle/supercomponent";
import {html, render} from "lit-html";
import env from "~brixi/controllers/env";

interface IDecksPage{}
export default class DecksPage extends SuperComponent<IDecksPage>{
    constructor(){
        super();
    }
    async connected(){
        await env.css(["decks-page"]);
        this.render();
    }
    render(){
        const view = html`Decks page`;
        render(view, this);
    }
}
env.bind("decks-page", DecksPage);
