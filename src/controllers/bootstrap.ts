import { mount } from "@codewithkyle/router";

(async ()=>{
    const loadingTextEl = document.body.querySelector("#loading-text");

    // Router
    loadingTextEl.innerHTML = "Loading application data.";
    // @ts-ignore
    await import("/js/routes.js");

    // Database
    loadingTextEl.innerHTML = "Launching local database.";
    // @ts-ignore
    const dbModule = await import("/js/database.js");
    await dbModule.default();

    // Starts app
    mount(document.body);
})();
