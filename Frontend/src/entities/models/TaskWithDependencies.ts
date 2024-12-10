import { Task } from './Task';

export interface TaskWithDependencies {
	id: string;
	title: string;
	description: string;
	status: string;
	project: string;
	member: string;
	blocked: boolean;
	dependencies: Task[];
}
