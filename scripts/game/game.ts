import { connect, send } from "../ws";

connect().then(() => {
    console.log("connected");
    setTimeout(()=>{
        const routeSegments = window.location.pathname.split("/");
        const gameID = routeSegments[routeSegments.length - 1].toLowerCase();
        console.log(gameID);
        if (gameID === "new"){
            send("hello world");
        } else {
            send(`dd::core::JOIN_ROOM::${gameID}`);
        }
    }, 5000);
});
