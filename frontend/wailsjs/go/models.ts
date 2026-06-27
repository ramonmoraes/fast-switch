export namespace main {
	
	export class AppState {
	    name: string;
	    platform: string;
	    capabilities: string[];
	    nextSteps: string[];
	
	    static createFrom(source: any = {}) {
	        return new AppState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.platform = source["platform"];
	        this.capabilities = source["capabilities"];
	        this.nextSteps = source["nextSteps"];
	    }
	}

}

