import env from "~brixi/controllers/env";
import { render, html } from "lit-html";

export default class MissingPage extends HTMLElement{
    constructor(){
        super();
        this.render();
    }
    render(){
        const view = html`404 | Page not found.`;
        render(view, this);
    }
}
env.bind("missing-page", MissingPage);
