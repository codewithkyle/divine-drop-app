import db from "@codewithkyle/jsql";
import SuperComponent from "@codewithkyle/supercomponent";
import {html, render} from "lit-html";
import env from "~brixi/controllers/env";

interface ICardModal {
    front: string,
    back: string,
    name: string,
}
export default class CardModal extends SuperComponent<ICardModal>{
    private cardId: string;

    constructor(cardId:string){
        super();
        this.cardId = cardId;
        this.model = {
            front: null,
            back: null,
            name: "",
        };
    }

    override async connected(){
        await env.css(["card-modal"]);
        const card = (await db.query<any>("SELECT front, back, name FROM cards WHERE id = $id", { id: this.cardId }))[0];
        this.set({
            front: card.front,
            back: card.back,
            name: card.name,
        });
    }

    private handleClick = (e) => {
        this.remove();
    }

    override render(): void {
        const view = html`
            <div class="backdrop" tabindex="0" @click=${this.handleClick}></div>
            <img src="${this.model.front}">
        `;
        render(view, this);
    }
}
env.bind("card-modal", CardModal);
