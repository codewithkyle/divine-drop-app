import Sortable from 'sortablejs';

declare const htmx:any;

class GroupedDeck extends HTMLElement {

    constructor() {
        super();
    }

    connectedCallback() {
        const el = this.querySelector('div[x-show]');
        new Sortable(el, {
            group: 'decks',
            sort: false,
            animation: 150,
            onEnd: (e) => {
                if (!e?.pullMode) return;
                const deck = e.item;
                const activeDeckId = document.body.querySelector<HTMLInputElement>('[name="active-deck-id"]')?.value ?? "";
                htmx.ajax('DELETE', `/decks/${deck.dataset.id}/group?active-deck-id=${activeDeckId}`, { target: '#decks', swap: 'none' });
            }
        });
    }
}
if (!customElements.get("deck-group")) customElements.define("deck-group", GroupedDeck);

class UngroupedDeck extends HTMLElement {

    constructor() {
        super();
    }

    connectedCallback() {
        new Sortable(this, {
            group: 'decks',
            sort: false,
            animation: 150,
            onEnd: (e) => {
                if (!e?.pullMode) return;
                const deck = e.item;
                const group = e.to.closest('deck-group');
                const activeDeckId = document.body.querySelector<HTMLInputElement>('[name="active-deck-id"]')?.value ?? "";
                htmx.ajax('PUT', `/groups/${group.dataset.id}/${deck.dataset.id}?active-deck-id=${activeDeckId}`, { target: '#decks', swap: 'none' });
            }
        });
    }
}
if (!customElements.get("ungrouped-decks")) customElements.define("ungrouped-decks", UngroupedDeck);
