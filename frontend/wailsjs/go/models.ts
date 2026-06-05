export namespace models {
	
	export class Availability {
	    id: number;
	    employeeId: number;
	    dayOfWeek: number;
	    startTime: string;
	    endTime: string;
	
	    static createFrom(source: any = {}) {
	        return new Availability(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.employeeId = source["employeeId"];
	        this.dayOfWeek = source["dayOfWeek"];
	        this.startTime = source["startTime"];
	        this.endTime = source["endTime"];
	    }
	}
	export class DaySettings {
	    dayOfWeek: number;
	    openTime: string;
	    closeTime: string;
	    schedulableOpen: string;
	    schedulableClose: string;
	    employeesNeeded: number;
	
	    static createFrom(source: any = {}) {
	        return new DaySettings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.dayOfWeek = source["dayOfWeek"];
	        this.openTime = source["openTime"];
	        this.closeTime = source["closeTime"];
	        this.schedulableOpen = source["schedulableOpen"];
	        this.schedulableClose = source["schedulableClose"];
	        this.employeesNeeded = source["employeesNeeded"];
	    }
	}
	export class Employee {
	    id: number;
	    name: string;
	    role: string;
	    desiredHours: number;
	    maxHours: number;
	    wage: number;
	
	    static createFrom(source: any = {}) {
	        return new Employee(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.name = source["name"];
	        this.role = source["role"];
	        this.desiredHours = source["desiredHours"];
	        this.maxHours = source["maxHours"];
	        this.wage = source["wage"];
	    }
	}
	export class Schedule {
	    id: number;
	    weekOf: string;
	    notes: string;
	
	    static createFrom(source: any = {}) {
	        return new Schedule(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.weekOf = source["weekOf"];
	        this.notes = source["notes"];
	    }
	}
	export class Settings {
	    id: number;
	    minShiftLength: number;
	    maxShiftLength: number;
	    allowOutsideHours: boolean;
	    timeBetweenShifts: number;
	
	    static createFrom(source: any = {}) {
	        return new Settings(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.minShiftLength = source["minShiftLength"];
	        this.maxShiftLength = source["maxShiftLength"];
	        this.allowOutsideHours = source["allowOutsideHours"];
	        this.timeBetweenShifts = source["timeBetweenShifts"];
	    }
	}
	export class Shift {
	    id: number;
	    scheduleId: number;
	    employeeId: number;
	    dayOfWeek: number;
	    startTime: string;
	    endTime: string;
	
	    static createFrom(source: any = {}) {
	        return new Shift(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.id = source["id"];
	        this.scheduleId = source["scheduleId"];
	        this.employeeId = source["employeeId"];
	        this.dayOfWeek = source["dayOfWeek"];
	        this.startTime = source["startTime"];
	        this.endTime = source["endTime"];
	    }
	}
	export class ValidationResult {
	    Errors: string[];
	    Warnings: string[];
	    Fatal: any;
	
	    static createFrom(source: any = {}) {
	        return new ValidationResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.Errors = source["Errors"];
	        this.Warnings = source["Warnings"];
	        this.Fatal = source["Fatal"];
	    }
	}

}

export namespace solver {
	
	export class ScoreResult {
	    hoursGap: number;
	    fairness: number;
	    clopenPenalty: number;
	    total: number;
	
	    static createFrom(source: any = {}) {
	        return new ScoreResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.hoursGap = source["hoursGap"];
	        this.fairness = source["fairness"];
	        this.clopenPenalty = source["clopenPenalty"];
	        this.total = source["total"];
	    }
	}
	export class SolverResult {
	    shifts: models.Shift[];
	    score: ScoreResult;
	    callCount: number;
	    wipeouts: number;
	    elapsed: number;
	    feasible: boolean;
	    solved: boolean;
	    timedOut: boolean;
	    message: string;
	
	    static createFrom(source: any = {}) {
	        return new SolverResult(source);
	    }
	
	    constructor(source: any = {}) {
	        if ('string' === typeof source) source = JSON.parse(source);
	        this.shifts = this.convertValues(source["shifts"], models.Shift);
	        this.score = this.convertValues(source["score"], ScoreResult);
	        this.callCount = source["callCount"];
	        this.wipeouts = source["wipeouts"];
	        this.elapsed = source["elapsed"];
	        this.feasible = source["feasible"];
	        this.solved = source["solved"];
	        this.timedOut = source["timedOut"];
	        this.message = source["message"];
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

