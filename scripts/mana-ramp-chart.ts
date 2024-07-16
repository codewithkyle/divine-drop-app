declare const Chart:any;

class ManaRampChart extends HTMLElement {
    private canvas: HTMLCanvasElement;

    constructor(){
        super();
    }

    connectedCallback(){
        this.canvas = document.createElement("canvas");
        this.appendChild(this.canvas);
        const ctx = this.canvas.getContext("2d");

        const range = Array.from({length: 11}, (_, i) => i);
        const cardsTMC = JSON.parse(this.dataset.tmcCards);

        let creatures = [];
        let enchantments = [];
        let sorceries = [];
        let artifacts = [];
        let instants = [];
        let allCards = [];

        for (let i = 0; i < range.length; i++) {
            if (i in cardsTMC) {
                allCards.push(cardsTMC[i].Count);
                creatures.push(cardsTMC[i].CreatureCount);
                enchantments.push(cardsTMC[i].EnchantmentCount);
                sorceries.push(cardsTMC[i].SorceryCount);
                artifacts.push(cardsTMC[i].ArtifactCount);
                instants.push(cardsTMC[i].InstantCount);
            } else {
                creatures.push(0);
                enchantments.push(0);
                sorceries.push(0);
                artifacts.push(0);
                instants.push(0);
                allCards.push(0);
            }
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
