export namespace desktop {
	
	export class ProfileSelection {
	    name: string;
	    isChina: boolean;
	    checked: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProfileSelection(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.isChina = source["isChina"];
	        this.checked = source["checked"];
	    }
	}
	export class ProfileSummary {
	    name: string;
	    isChina: boolean;
	
	    static createFrom(source: any = {}) {
	        return new ProfileSummary(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.isChina = source["isChina"];
	    }
	}
	export class RotationRequest {
	    profiles: ProfileSelection[];
	    username: string;
	    password: string;
	    mode: string;
	
	    static createFrom(source: any = {}) {
	        return new RotationRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.profiles = this.convertValues(source["profiles"], ProfileSelection);
	        this.username = source["username"];
	        this.password = source["password"];
	        this.mode = source["mode"];
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

export namespace rotator {
	
	export class ProfileResult {
	    name: string;
	    success: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new ProfileResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.name = source["name"];
	        this.success = source["success"];
	        this.message = source["message"];
	    }
	}

}

