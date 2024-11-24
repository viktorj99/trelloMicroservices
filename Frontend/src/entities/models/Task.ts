export interface Task {
    id?: string;
    title: string;
    description?: string;
    status: "PENDING" | "IN_PROGRESS" | "FINISHED"; 
    project: string;
    member: string;
    blocked?: boolean;
}