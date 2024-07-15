declare const Chart:any;

class ManaRampChart extends HTMLElement {
    private canvas: HTMLCanvasElement;

    constructor(){
        super();
    }

    connectedCallback(){
        this.canvas = this.querySelector("canvas");
        const ctx = this.canvas.getContext("2d");

        //const range = JSON.parse(this.dataset.tmcRange);
        const range = Array.from({length: 11}, (_, i) => i);
        const tmc = JSON.parse(this.dataset.tmc)
        const tmcCreature = JSON.parse(this.dataset.tmcCreature);
        const tmcArtifact = JSON.parse(this.dataset.tmcArtifact);
        const tmcEnchantment = JSON.parse(this.dataset.tmcEnchantment);
        const tmcSorcery = JSON.parse(this.dataset.tmcSorcery);
        const tmcInstant = JSON.parse(this.dataset.tmcInstant);

        let creatures = [];
        let enchantments = [];
        let sorceries = [];
        let artifacts = [];
        let instants = [];
        let allCards = [];

        for (let i = 0; i < range.length; i++) {
            let foundCard = false;
            for (let j = 0; j < tmc.length; j++) {
                if (tmc[j].TMC === range[i]) {
                    allCards.push(tmc[j].Count);
                    foundCard = true;
                    break;
                }
            }
            if (!foundCard) allCards.push(0);

            let foundCreature = false;
            for (let j = 0; j < tmcCreature.length; j++) {
                if (tmcCreature[j].TMC === range[i]) {
                    creatures.push(tmcCreature[j].Count);
                    foundCreature = true;
                    break;
                }
            }
            if (!foundCreature) creatures.push(0);

            let foundArtifact = false;
            for (let j = 0; j < tmcArtifact.length; j++) {
                if (tmcArtifact[j].TMC === range[i]) {
                    artifacts.push(tmcArtifact[j].Count);
                    foundArtifact = true;
                    break;
                }
            }
            if (!foundArtifact) artifacts.push(0);

            let foundEnchantment = false;
            for (let j = 0; j < tmcEnchantment.length; j++) {
                if (tmcEnchantment[j].TMC === range[i]) {
                    enchantments.push(tmcEnchantment[j].Count);
                    foundEnchantment = true;
                    break;
                }
            }
            if (!foundEnchantment) enchantments.push(0);

            let foundSorcery = false;
            for (let j = 0; j < tmcSorcery.length; j++) {
                if (tmcSorcery[j].TMC === range[i]) {
                    sorceries.push(tmcSorcery[j].Count);
                    foundSorcery = true;
                    break;
                }
            }
            if (!foundSorcery) sorceries.push(0);

            let foundInstant = false;
            for (let j = 0; j < tmcInstant.length; j++) {
                if (tmcInstant[j].TMC === range[i]) {
                    instants.push(tmcInstant[j].Count);
                    foundInstant = true;
                    break;
                }
            }
            if (!foundInstant) instants.push(0);
        }

        new Chart(ctx, {
            type: 'line',
            data: {
                labels: range,
                datasets: [
                    {
                        label: "All Cards",
                        data: allCards,
                        cubicInterpolationMode: 'monotone',
                        tension: 0.4,
                        fill: false,
                    },
                    {
                        label: "Creatures",
                        data: creatures,
                        cubicInterpolationMode: 'monotone',
                        tension: 0.4,
                        fill: false,
                        hidden: true,
                    },
                    {
                        label: "Artifacts",
                        data: artifacts,
                        cubicInterpolationMode: 'monotone',
                        tension: 0.4,
                        fill: false,
                        hidden: true,
                    },
                    {
                        label: "Enchantments",
                        data: enchantments,
                        cubicInterpolationMode: 'monotone',
                        tension: 0.4,
                        fill: false,
                        hidden: true,
                    },
                    {
                        label: "Sorceries",
                        data: sorceries,
                        cubicInterpolationMode: 'monotone',
                        tension: 0.4,
                        fill: false,
                        hidden: true,
                    },
                    {
                        label: "Instants",
                        data: instants,
                        cubicInterpolationMode: 'monotone',
                        tension: 0.4,
                        fill: false,
                        hidden: true,
                    },
                ]
            },
            options: {
                scales: {
                    y: {
                        title: {
                            display: true,
                            text: "Number of Cards",
                            color: "#ffffff",
                        },
                        ticks: {
                            color: "#ffffff",
                            stepSize: 1,
                        },
                        grid: {
                            color: "rgba(255,255,255,0.1)"
                        },
                    },
                    x: {
                        title: {
                            display: true,
                            text: "Total Mana Cost",
                            color: "#ffffff"
                        },
                        ticks: {
                            color: "#ffffff"
                        },
                        grid: {
                            color: "rgba(255,255,255,0.1)"
                        },
                    },
                },
                plugins: {
                    title: {
                        display: true,
                        text: "Mana Curve",
                        color: "#ffffff",
                    },
                    legend: {
                        labels: {
                            color: "#ffffff",
                        },
                    },
                },
            },
        });
    }
}
if (!customElements.get("mana-ramp-chart")) customElements.define("mana-ramp-chart", ManaRampChart);
