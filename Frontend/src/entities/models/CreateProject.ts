import { User } from "./User";

export interface  CreateProject {
	id?: string; 
  name: string;
  expectedEndDate?: string; 
  minMembers?: number;
  maxMembers?: number;
  manager?: User; 
  members?:string[];
  isDeleted?: boolean;
}

export interface DTOCreateProject {
	id?: string; 
  name: string;
  expectedEndDate?: string; 
  minMembers?: number;
  maxMembers?: number;
  manager?: User; 
  members?:User[];
  isDeleted?: boolean;
}

