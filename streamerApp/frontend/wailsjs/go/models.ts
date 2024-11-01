export namespace types {
	
	export class Config {
	    ObsHost: string;
	    ObsPort: string;
	    ObsPassword: string;
	    TwitchUsername: string;
	
	    static createFrom(source: any = {}) {
	        return new Config(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.ObsHost = source["ObsHost"];
	        this.ObsPort = source["ObsPort"];
	        this.ObsPassword = source["ObsPassword"];
	        this.TwitchUsername = source["TwitchUsername"];
	    }
	}

}

