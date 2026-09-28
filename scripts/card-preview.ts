class CardPreviewButton extends HTMLElement{
    private card: HTMLImageElement | null;

    constructor(){
        super();
        this.card = null;
    }

    connectedCallback(){
        this.addEventListener("mouseenter", this.onMouseEnter);
        this.addEventListener("focus", this.onMouseEnter);
        this.addEventListener("mouseleave", this.onMouseLeave);
        this.addEventListener("blur", this.onMouseLeave);
    }

    private onMouseEnter = () => {
        if (this.card && this.card.isConnected){
            this.card.remove();
        }
        this.card = document.createElement("img");
        this.card.src = this.dataset.cardUrl || "";
        if (!this.card.src) return;
        const bounds = this.getBoundingClientRect();
        const width = 350;
        const gap = 8;
        this.card.style.position = "fixed";
        let bottom = bounds.top;
        if (bottom + 488 > window.innerHeight){
            bottom = window.innerHeight - 488;
        }
        if (bottom < gap){
            bottom = gap;
        }
        // Prefer the left, which is where the deck tray sits. Flip to the right
        // when there is no room, so a trigger in a left hand column does not
        // preview off screen.
        let left = bounds.left - width;
        if (left < gap){
            left = bounds.right + gap;
        }
        if (left + width > window.innerWidth - gap){
            left = Math.max(gap, window.innerWidth - width - gap);
        }
        this.card.style.top = `${bottom}px`;
        this.card.style.left = `${left}px`;
        this.card.style.width = `${width}px`;
        this.card.style.boxShadow = "var(--shadow-black-lg)";
        this.card.style.borderRadius = "4%";
        this.card.style.zIndex = "1000";
        this.card.style.opacity = "0";
        this.card.style.transition = "opacity 150ms var(--ease-in-out)";
        this.card.style.pointerEvents = "none";
        this.card.addEventListener("load", () => {
            if (this.card){
                this.card.style.opacity = "1";
            }
        });
        document.body.appendChild(this.card);
    }

    private onMouseLeave = () => {
        if (this.card && this.card.isConnected){
            this.card.remove();
        }
    }
}
if (!customElements.get("card-preview")) customElements.define("card-preview", CardPreviewButton);
