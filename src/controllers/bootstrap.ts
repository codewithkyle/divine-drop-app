import db from "@codewithkyle/jsql";

const loadingTextEl = document.body.querySelector("#loading-text");

async function prepDatabase(){
    loadingTextEl.innerHTML = "Launching local database.";
    await db.start({
        cache: ["cards", "decks"],
        schema: {
            name: "app",
            version: 1,
            tables: [
                {
                    name: "cards",
                    keyPath: "id",
                    columns: [
                        {
                            key: "id",
                            unique: true,
                        },
                        {
                            key: "name",
                        },
                        {
                            key: "layout",
                        },
                        {
                            key: "colors",
                        },
                        {
                            key: "legalities",
                        },
                        {
                            key: "rarity",
                        },
                        {
                            key: "keywords",
                        },
                        {
                            key: "front",
                        },
                        {
                            key: "back",
                        },
                        {
                            key: "type",
                        },
                        {
                            key: "subtypes",
                        },
                        {
                            key: "texts",
                        },
                        {
                            key: "manaCosts",
                        },
                        {
                            key: "totalManaCost",
                        },
                        {
                            key: "faceNames",
                        },
                        {
                            key: "flavorTexts",
                        },
                        {
                            key: "toughness",
                        },
                        {
                            key: "power",
                        },
                    ],
                },
                {
                    name: "decks",
                    keyPath: "id",
                    persist: true,
                    columns: [
                        {
                            key: "id",
                            unique: true,
                        },
                        {
                            key: "label",
                            default: "Untitled",
                        },
                        {
                            key: "commanderId",
                        },
                        {
                            key: "cards",
                            default: [],
                        },
                        {
                            key: "dateCreated",
                        },
                        {
                            key: "dateUpdated",
                        }
                    ],
                },
            ],
        },
    });
    console.log("db started");

    const res = await fetch("https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards.jsonl", {
        method: "HEAD",
    });
    const lastModified = res.headers.get("Last-modified");
    const lastSeen = localStorage.getItem("db-timestamp");
    const cardCount = await db.query("SELECT COUNT(*) FROM cards");
    if (lastModified !== lastSeen || cardCount[0]["COUNT(*)"] < 25_000){
        const loadingTextEl = document.body.querySelector("#loading-text");
        loadingTextEl.innerHTML = `Downloading card data<br><span style="font-size:12px;color:var(--grey-500);">(this may take several minutes)</span>`;
        await db.query("RESET cards");
        await db.ingest("https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards.jsonl", "cards", "NDJSON");
        localStorage.setItem("db-timestamp", lastModified);
    }
    console.log("db ready");
    document.dispatchEvent(new CustomEvent("init:app"));
}

async function startRouter(){
    // Router
    loadingTextEl.innerHTML = "Loading application data.";
    
    // @ts-ignore
    import("/js/tooltipper.js");

    // @ts-ignore
    await import("/js/routes.js");

    // Starts app
    document.body.innerHTML = `
        <sidebar-component></sidebar-component>
        <main></main>
    `;
    // @ts-ignore
    await import("/js/sidebar-component.js");
    const main = document.body.querySelector("main");
    // @ts-ignore
    const { mount } = await import("/js/router.js");
    mount(main);
}

document.addEventListener("init:app", () => {
    startRouter();
});

prepDatabase();
