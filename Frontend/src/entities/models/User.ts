import { Role } from './Role';

export interface User {
	id?: string;
	username: string;
	role: Role;
}
