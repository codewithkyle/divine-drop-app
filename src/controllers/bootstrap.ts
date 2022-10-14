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
    document.body.innerHTML = `
        <sidebar-component></sidebar-component>
        <main></main>
    `;
    // @ts-ignore
    await import("/js/sidebar-component.js");
    const main = document.body.querySelector("main");
    mount(main);

    // @ts-ignore
    import("/js/tooltipper.js");
})();
