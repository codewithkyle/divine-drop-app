import toaster from "@codewithkyle/notifyjs/dist/toaster";

document.addEventListener("flash:toast", (event:CustomEvent) => {
    const message = event.detail.value;
    toaster.push({
        message,
        duration: 10,
    });
});
