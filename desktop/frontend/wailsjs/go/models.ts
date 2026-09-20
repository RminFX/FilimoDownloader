export namespace api {
	
	export class Episode {
	    id: string;
	    title: string;
	    season: number;
	    number: number;
	    duration: string;
	    seasonTitle: string;
	    cover: string;
	    poster: string;
	    description: string;
	
	    static createFrom(source: any = {}) {
	        return new Episode(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.season = source["season"];
	        this.number = source["number"];
	        this.duration = source["duration"];
	        this.seasonTitle = source["seasonTitle"];
	        this.cover = source["cover"];
	        this.poster = source["poster"];
	        this.description = source["description"];
	    }
	}

}

export namespace engine {
	
	export class QueueItem {
	    id: string;
	    title: string;
	    kind: string;
	    seriesKey: string;
	    seriesTitle: string;
	    season: number;
	    episode: number;
	    seasonTitle: string;
	    cover: string;
	
	    static createFrom(source: any = {}) {
	        return new QueueItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.kind = source["kind"];
	        this.seriesKey = source["seriesKey"];
	        this.seriesTitle = source["seriesTitle"];
	        this.season = source["season"];
	        this.episode = source["episode"];
	        this.seasonTitle = source["seasonTitle"];
	        this.cover = source["cover"];
	    }
	}
	export class EnqueueRequest {
	    items: QueueItem[];
	    quality: string;
	    audio: string[];
	    subtitles: string[];
	    format: string;
	
	    static createFrom(source: any = {}) {
	        return new EnqueueRequest(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.items = this.convertValues(source["items"], QueueItem);
	        this.quality = source["quality"];
	        this.audio = source["audio"];
	        this.subtitles = source["subtitles"];
	        this.format = source["format"];
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
	export class HistoryItem {
	    id: string;
	    title: string;
	    quality: string;
	    format: string;
	    sizeMb: number;
	    downloadedAt: string;
	    filePath: string;
	
	    static createFrom(source: any = {}) {
	        return new HistoryItem(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.quality = source["quality"];
	        this.format = source["format"];
	        this.sizeMb = source["sizeMb"];
	        this.downloadedAt = source["downloadedAt"];
	        this.filePath = source["filePath"];
	    }
	}
	export class Quality {
	    quality: string;
	    resolution: string;
	
	    static createFrom(source: any = {}) {
	        return new Quality(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.quality = source["quality"];
	        this.resolution = source["resolution"];
	    }
	}
	export class InspectResult {
	    id: string;
	    title: string;
	    kind: string;
	    season: number;
	    episode: number;
	    cover: string;
	    description: string;
	    categories: string[];
	    qualities: Quality[];
	    audio: string[];
	    subtitles: string[];
	    episodes: api.Episode[];
	    seasonIncomplete: boolean;
	    seriesTitle: string;
	    parentId: string;
	
	    static createFrom(source: any = {}) {
	        return new InspectResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.title = source["title"];
	        this.kind = source["kind"];
	        this.season = source["season"];
	        this.episode = source["episode"];
	        this.cover = source["cover"];
	        this.description = source["description"];
	        this.categories = source["categories"];
	        this.qualities = this.convertValues(source["qualities"], Quality);
	        this.audio = source["audio"];
	        this.subtitles = source["subtitles"];
	        this.episodes = this.convertValues(source["episodes"], api.Episode);
	        this.seasonIncomplete = source["seasonIncomplete"];
	        this.seriesTitle = source["seriesTitle"];
	        this.parentId = source["parentId"];
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
	export class Job {
	    id: string;
	    contentId: string;
	    title: string;
	    quality: string;
	    format: string;
	    status: string;
	    progress: number;
	    message: string;
	    filePath: string;
	    error: string;
	    kind: string;
	    seriesKey: string;
	    seriesTitle: string;
	    season: number;
	    episode: number;
	    seasonTitle: string;
	    cover: string;
	    audio: string[];
	    subs: string[];
	
	    static createFrom(source: any = {}) {
	        return new Job(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.contentId = source["contentId"];
	        this.title = source["title"];
	        this.quality = source["quality"];
	        this.format = source["format"];
	        this.status = source["status"];
	        this.progress = source["progress"];
	        this.message = source["message"];
	        this.filePath = source["filePath"];
	        this.error = source["error"];
	        this.kind = source["kind"];
	        this.seriesKey = source["seriesKey"];
	        this.seriesTitle = source["seriesTitle"];
	        this.season = source["season"];
	        this.episode = source["episode"];
	        this.seasonTitle = source["seasonTitle"];
	        this.cover = source["cover"];
	        this.audio = source["audio"];
	        this.subs = source["subs"];
	    }
	}
	export class LibrarySeason {
	    seriesKey: string;
	    season: number;
	    title: string;
	    cover: string;
	    episodeCount: number;
	    doneCount: number;
	    progress: number;
	
	    static createFrom(source: any = {}) {
	        return new LibrarySeason(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.seriesKey = source["seriesKey"];
	        this.season = source["season"];
	        this.title = source["title"];
	        this.cover = source["cover"];
	        this.episodeCount = source["episodeCount"];
	        this.doneCount = source["doneCount"];
	        this.progress = source["progress"];
	    }
	}
	export class LibrarySeries {
	    key: string;
	    title: string;
	    cover: string;
	    kind: string;
	    seasonCount: number;
	    episodeCount: number;
	    doneCount: number;
	    activeCount: number;
	    progress: number;
	
	    static createFrom(source: any = {}) {
	        return new LibrarySeries(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.key = source["key"];
	        this.title = source["title"];
	        this.cover = source["cover"];
	        this.kind = source["kind"];
	        this.seasonCount = source["seasonCount"];
	        this.episodeCount = source["episodeCount"];
	        this.doneCount = source["doneCount"];
	        this.activeCount = source["activeCount"];
	        this.progress = source["progress"];
	    }
	}
	
	
	export class Session {
	    loggedIn: boolean;
	    username: string;
	    hasToken: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Session(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.loggedIn = source["loggedIn"];
	        this.username = source["username"];
	        this.hasToken = source["hasToken"];
	    }
	}
	export class Settings {
	    defaultQuality: string;
	    defaultFormat: string;
	    downloadPath: string;
	    autoOpenFolder: boolean;
	    showInfoBeforeDL: boolean;
	    concurrentDownloads: number;
	    theme: string;
	    language: string;
	    notifications: boolean;
	    checkUpdatesOnLaunch: boolean;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.defaultQuality = source["defaultQuality"];
	        this.defaultFormat = source["defaultFormat"];
	        this.downloadPath = source["downloadPath"];
	        this.autoOpenFolder = source["autoOpenFolder"];
	        this.showInfoBeforeDL = source["showInfoBeforeDL"];
	        this.concurrentDownloads = source["concurrentDownloads"];
	        this.theme = source["theme"];
	        this.language = source["language"];
	        this.notifications = source["notifications"];
	        this.checkUpdatesOnLaunch = source["checkUpdatesOnLaunch"];
	    }
	}
	export class UpdateInfo {
	    current: string;
	    latest: string;
	    url: string;
	    available: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new UpdateInfo(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.current = source["current"];
	        this.latest = source["latest"];
	        this.url = source["url"];
	        this.available = source["available"];
	        this.message = source["message"];
	    }
	}

}

