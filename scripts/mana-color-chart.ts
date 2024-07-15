class ManaColorChart extends HTMLElement {
    private canvas: HTMLCanvasElement;

    constructor(){
        super();
    }

    connectedCallback(){
        this.canvas = document.createElement("canvas");
        this.appendChild(this.canvas);
        const ctx = this.canvas.getContext("2d");

        const range = ["White","Blue","Black","Red","Green"];

        const colorCounts = JSON.parse(this.dataset.colorCounts);
        const creatureCount = JSON.parse(this.dataset.creatureColorCounts);
        const enchantmentCount = JSON.parse(this.dataset.enchantmentColorCounts);
        const instantCount = JSON.parse(this.dataset.instantColorCounts);
        const sorceryCount = JSON.parse(this.dataset.sorceryColorCounts);

        new Chart(ctx, {
            type: 'radar',
            data: {
                labels: range,
                datasets: [
                    {
                        label: 'All Cards',
                        data: colorCounts,
                    },
                    {
                        label: 'Creatures',
                        data: creatureCount,
                        hidden: true,
                    },
                    {
                        label: 'Enchantments',
                        data: enchantmentCount,
                        hidden: true,
                    },
                    {
                        label: 'Instants',
                        data: instantCount,
                        hidden: true,
                    },
                    {
                        label: 'Sorceries',
                        data: sorceryCount,
                        hidden: true,
                    },
                ]
            },
            options: {
                spanGaps: true,
                scales: {
                    r: {
                        angleLines: {
                            color: 'rgba(255, 255, 255, 0.2)',
                        },
                        grid: {
                            color: 'rgba(255, 255, 255, 0.2)',
                        },
                        pointLabels: {
                            color: '#FFFFFF',
                        },
                        ticks: {
                            backdropColor: 'rgba(0, 0, 0, 0.5)',
                            color: '#FFFFFF',
                            stepSize: 1,
                            display: true,
                            beginAtZero: true,
                            min: 0,
                        }
                    }
                },
                plugins: {
                    title: {
                        display: true,
                        text: "Devotion",
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
if (!customElements.get("mana-color-chart")) customElements.define("mana-color-chart", ManaColorChart);
