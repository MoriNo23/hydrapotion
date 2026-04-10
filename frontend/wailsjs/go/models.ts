export namespace main {
	
	export class HistoryDay {
	    day: string;
	    ml: number;
	    date: string;
	
	    static createFrom(source: any = {}) {
	        return new HistoryDay(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.day = source["day"];
	        this.ml = source["ml"];
	        this.date = source["date"];
	    }
	}
	export class MoodEntry {
	    date: string;
	    mood: string;
	    ml: number;
	
	    static createFrom(source: any = {}) {
	        return new MoodEntry(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.date = source["date"];
	        this.mood = source["mood"];
	        this.ml = source["ml"];
	    }
	}
	export class Settings {
	    weight: number;
	    today_consumed: number;
	    daily_goal: number;
	    language: string;
	    location: string;
	    last_reset_date: string;
	    reminder_interval: number;
	    current_mood: string;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.weight = source["weight"];
	        this.today_consumed = source["today_consumed"];
	        this.daily_goal = source["daily_goal"];
	        this.language = source["language"];
	        this.location = source["location"];
	        this.last_reset_date = source["last_reset_date"];
	        this.reminder_interval = source["reminder_interval"];
	        this.current_mood = source["current_mood"];
	    }
	}

}

