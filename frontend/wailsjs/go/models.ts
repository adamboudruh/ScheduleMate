export namespace models {
	
	export class Employee {
	    id: number;
	    name: string;
	    role: string;
	    desiredHours: number;
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
	        this.wage = source["wage"];
	    }
	}

}

