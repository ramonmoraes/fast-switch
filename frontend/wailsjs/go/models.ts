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
	export class SwitcherState {
	    visible: boolean;
	    selectedIndex: number;
	    selectedWindow?: WindowInfo;
	
	    static createFrom(source: any = {}) {
	        return new SwitcherState(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.visible = source["visible"];
	        this.selectedIndex = source["selectedIndex"];
	        this.selectedWindow = this.convertValues(source["selectedWindow"], WindowInfo);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	export class WindowInfo {
	    ownerName: string;
	    title: string;
	    icon: string;
	    pid: number;
	    layer: number;
	    x: number;
	    y: number;
	    width: number;
	    height: number;
	
	    static createFrom(source: any = {}) {
	        return new WindowInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ownerName = source["ownerName"];
	        this.title = source["title"];
	        this.icon = source["icon"];
	        this.pid = source["pid"];
	        this.layer = source["layer"];
	        this.x = source["x"];
	        this.y = source["y"];
	        this.width = source["width"];
	        this.height = source["height"];
	    }
	}
	export class PermissionStatus {
	    accessibility: boolean;
	    screenRecording: boolean;
	    hotkeyRegistered: boolean;
	    hotkeyPresses: number;
	    statusItemReady: boolean;
	    warnings: string[];
	
	    static createFrom(source: any = {}) {
	        return new PermissionStatus(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.accessibility = source["accessibility"];
	        this.screenRecording = source["screenRecording"];
	        this.hotkeyRegistered = source["hotkeyRegistered"];
	        this.hotkeyPresses = source["hotkeyPresses"];
	        this.statusItemReady = source["statusItemReady"];
	        this.warnings = source["warnings"];
	    }
	}
	export class DesktopSnapshot {
	    appState: AppState;
	    permissions: PermissionStatus;
	    windows: WindowInfo[];
	    switcher: SwitcherState;
	
	    static createFrom(source: any = {}) {
	        return new DesktopSnapshot(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.appState = this.convertValues(source["appState"], AppState);
	        this.permissions = this.convertValues(source["permissions"], PermissionStatus);
	        this.windows = this.convertValues(source["windows"], WindowInfo);
	        this.switcher = this.convertValues(source["switcher"], SwitcherState);
	    }
	
		convertValues(a: any, classs: any, asMap: boolean = false): any {
		    if (!a) {
		        return a;
		    }
		    if (a.slice && a.map) {
		        return (a as any[]).map(elem => this.convertValues(elem, classs));
		    } else if ("object" === typeof a) {
		        if (asMap) {
		            for (const key of Object.keys(a)) {
		                a[key] = new classs(a[key]);
		            }
		            return a;
		        }
		        return new classs(a);
		    }
		    return a;
		}
	}
	
	

}

