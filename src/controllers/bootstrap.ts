import { router, mount } from "@codewithkyle/router";
import db from "@codewithkyle/jsql";

(async ()=>{
    // Router
    router.redirect("/", "/decks");
    router.add("/decks", "decks-page");
    router.add("/*", "missing-page");

    // Database
    await db.start({
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

    const res = await fetch("https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards.jsonl", {
        method: "HEAD",
    });
    const lastModified = res.headers.get("Last-modified");
    const lastSeen = localStorage.getItem("db-timestamp");
    if (lastModified !== lastSeen){
        const loadingTextEl = document.body.querySelector("#loading-text");
        loadingTextEl.innerHTML = "Loading card data...";
        await db.ingest("https://divinedrop.nyc3.cdn.digitaloceanspaces.com/cards.jsonl", "cards", "NDJSON");
        localStorage.setItem("db-timestamp", lastModified);
    }

    // Start app
    mount(document.body);
})();
