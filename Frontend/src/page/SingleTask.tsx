import { useMutation, useQuery, useQueryClient } from '@tanstack/react-query';
import { useParams } from 'react-router-dom';
import { getTaskById, getTasksByProjectId } from '../services/taskService';
import { Button, notification, Table } from 'antd';
import { addTaskDependency } from '../services/workflowService';
import { getTasksWithDependencies } from '../services/taskService';

const SingleTask = () => {
	const queryClient = useQueryClient();
	const { taskId, projectId } = useParams<{ taskId: string; projectId: string }>();
	const { data: task } = useQuery({
		queryKey: ['task', taskId],
		queryFn: () => getTaskById(taskId!),
	});

	const { data: tasks } = useQuery({
		queryKey: ['tasks', projectId],
		queryFn: () => getTasksByProjectId(projectId!),
	});

	const { data: tasksWithDependencies } = useQuery({
		queryKey: ['tasksWithDependencies'],
		queryFn: () => getTasksWithDependencies(projectId!),
	});

	const filteredTasks = tasks?.filter((t) => t.id !== taskId);
	const mutation = useMutation({
		mutationFn: (data: { taskId: string; dependencyId: string }) =>
			addTaskDependency(data.taskId, data.dependencyId),
		onSuccess: () => {
			queryClient.invalidateQueries({ queryKey: ['task', taskId] });
			queryClient.invalidateQueries({ queryKey: ['tasks', projectId] });
			queryClient.invalidateQueries({ queryKey: ['tasksWithDependencies'] });
			notification.success({
				message: 'Success',
				description: 'Task dependency added successfully',
			});
		},
		onError: (error) => {
			notification.error({
				message: 'Error',
				description:
					error instanceof Error ? error.message : 'Failed to add task dependency',
			});
		},
	});

	const hasDependency = (taskIdFunc: string) => {
		if (!tasksWithDependencies) return false;
		const taskMain = tasksWithDependencies?.find((t) => t.id === taskId);
		console.log(taskMain);
		return taskMain?.dependencies?.some(
			(dep) => dep.id === tasks?.find((t) => t.id === taskIdFunc)?.id
		);
		// return tasksWithDependencies.some((task) =>
		// 	task.dependencies?.some((dep) => dep.id === taskId)
		// );
	};

	return (
		<div>
			<h1>{task?.title}</h1>
			<p>{task?.description}</p>
			<p>Status: {task?.status}</p>
			<p>Project: {projectId}</p>
			<Table dataSource={filteredTasks} rowKey='id'>
				<Table.Column title='Title' dataIndex='title' />
				<Table.Column title='Description' dataIndex='description' />
				<Table.Column title='Status' dataIndex='status' />
				<Table.Column title='Member' dataIndex='member' />
				<Table.Column
					title='Action'
					render={(record) => (
						<Button
							disabled={hasDependency(record.id)}
							onClick={() =>
								mutation.mutate({
									taskId: taskId!,
									dependencyId: record.id,
								})
							}
						>
							Add Dependency
						</Button>
					)}
				/>
			</Table>
		</div>
	);
};

export default SingleTask;
