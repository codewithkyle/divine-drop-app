declare const htmx:any;

class GroupedDeck extends HTMLElement {

    private dragging: boolean;

    constructor() {
        super();
         this.dragging = false
    }

    connectedCallback() {
        this.addEventListener("dragover", this.onDragEnter);
        this.addEventListener("dragleave", this.onDragLeave);
        this.addEventListener("drop", this.onDrop, { capture: true });
        this.addEventListener("dragend", this.onDragEnd);
        this.addEventListener("dragstart", this.onDragStart);

        this.querySelectorAll(".deck-link").forEach(el => {
            el.addEventListener("dragstart", (e: DragEvent) => {
                this.dragging = true;
                const target = e.currentTarget as HTMLElement;
                e.dataTransfer.setData('text/plain', target.dataset.id);
            });
        });
    }

    onDragStart:EventListener = (e) => {
        this.querySelectorAll(".deck-link").forEach((el:HTMLElement) => {
            el.classList.add("disabled");
        });
    }

    onDragEnd:EventListener = (e) => {
        e.preventDefault();
        this.classList.remove("dropzone");
        this.querySelectorAll(".deck-link").forEach((el:HTMLElement) => {
            el.classList.remove("disabled");
        });
    }

    onDrop:EventListener = (e:DragEvent) => {
        e.preventDefault();
        if (this.dragging) {
            this.dragging = false;
            return;
        }
        this.dragging = false;
        const id = e.dataTransfer.getData('text/plain');
        const activeDeckId = document.body.querySelector<HTMLInputElement>('[name="active-deck-id"]')?.value ?? "";
        htmx.ajax('PUT', `/groups/${this.dataset.id}/${id}?active-deck-id=${activeDeckId}`, { target: '#decks', swap: 'innerHTML' });
    }

    onDragEnter:EventListener = (e) => {
        e.preventDefault();
        this.classList.add("dropzone");
    }

    onDragLeave:EventListener = (e) => {
        e.preventDefault();
        this.classList.remove("dropzone");
    }
}
if (!customElements.get("deck-group")) customElements.define("deck-group", GroupedDeck);

class UngroupedDeck extends HTMLElement {

    private dragging:boolean;

    constructor() {
        super();
        this.dragging = false;
    }

    connectedCallback() {
        this.addEventListener("dragover", this.onDragEnter);
        this.addEventListener("dragleave", this.onDragLeave);
        this.addEventListener("drop", this.onDrop, { capture: true });
        this.addEventListener("dragend", this.onDragEnd);
        this.addEventListener("dragstart", this.onDragStart);

        this.querySelectorAll(".deck-link").forEach(el => {
            el.addEventListener("dragstart", (e: DragEvent) => {
                this.dragging = true;
                const target = e.currentTarget as HTMLElement;
                e.dataTransfer.setData('text/plain', target.dataset.id);
            });
        });
    }

    onDragStart:EventListener = (e) => {
        this.querySelectorAll(".deck-link").forEach((el:HTMLElement) => {
            el.classList.add("disabled");
        });
    }

    onDragEnd:EventListener = (e) => {
        e.preventDefault();
        this.classList.remove("dropzone");
        this.querySelectorAll(".deck-link").forEach((el:HTMLElement) => {
            el.classList.remove("disabled");
        });
    }

    onDrop:EventListener = (e:DragEvent) => {
        e.preventDefault();
        if (this.dragging) {
            this.dragging = false;
            return;
        }
        this.dragging = false;
        const id = e.dataTransfer.getData('text/plain');
        const activeDeckId = document.body.querySelector<HTMLInputElement>('[name="active-deck-id"]')?.value ?? "";
        htmx.ajax('DELETE', `/decks/${id}/group?active-deck-id=${activeDeckId}`, { target: '#decks', swap: 'innerHTML' });
    }

    onDragEnter:EventListener = (e) => {
        e.preventDefault();
        this.classList.add("dropzone");
    }

    onDragLeave:EventListener = (e) => {
        e.preventDefault();
        this.classList.remove("dropzone");
    }
}
if (!customElements.get("ungrouped-decks")) customElements.define("ungrouped-decks", UngroupedDeck);
