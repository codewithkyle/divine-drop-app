export function parse(str:string): string{
    const segments = str.match(/\{.*?\}/g);
    for (let i = 0; i < segments.length; i++){
        const symbol = segments[i].replace(/\{|\}/g, "");
        const url = `https://divinedrop.nyc3.cdn.digitaloceanspaces.com/symbols/${symbol}.svg`;
        str = str.replace(segments[i], `<img src=\"${url}\">`);
    }
    return str;
}
