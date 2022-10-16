import { publish } from "@codewithkyle/pubsub";

class Editor {
    private query: string;
    private sort: string;
    private colors: {
        white: boolean,
        black: boolean,
        blue: boolean,
        red: boolean,
        green: boolean,
    }
    private type: string;
    private subtypes: string[];
    private legality: string;

    constructor(){
        this.query = "";
        this.sort = "name";
        this.colors = {
            white: false,
            black: false,
            blue: false,
            red: false,
            green: false,
        }
        this.type = null;
        this.subtypes = [];
        this.legality = null;
    }

    private dispatch(){
        publish("deck-editor", {
            colors: this.colors,
            query: this.query,
            sort: this.sort,
            type: this.type,
            subtypes: [...this.subtypes],
            legality: this.legality,
        });
    }

    public setLegality(mode){
        this.legality = mode;
        this.dispatch();
    }

    public setQuery(value:string):void{
        this.query = value;
        this.dispatch();
    }

    public setSort(value:string):void{
        this.sort = value;
        this.dispatch();
    }

    public setColor(color:string, value:boolean):void{
        this.colors[color] = value;
        this.dispatch();
    }

    public setType(type:string):void{
        this.type = type;
        this.dispatch();
    }

    public addSubtype(type:string):void{
        this.subtypes.push(type);
        this.subtypes = [...new Set(this.subtypes)];
        this.dispatch();
    }

    public removeSubtype(value:string):void{
        const index = this.subtypes.indexOf(value);
        this.subtypes.splice(index, 1);
        this.dispatch();
    }
}
const editor = new Editor();
export default editor;
