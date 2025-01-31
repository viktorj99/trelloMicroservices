import React, { useState, useEffect } from 'react';
import { Table, notification } from 'antd';
import { getTaskStatusHistory } from '../services/taskService';
import { useParams } from 'react-router-dom';

const TaskStatusHistoryPage = () => {
	const { taskId } = useParams<{ taskId: string }>();
	const [statusHistory, setStatusHistory] = useState([]);
	const [loading, setLoading] = useState(false);

	useEffect(() => {
		const fetchTaskStatusHistory = async () => {
			try {
				setLoading(true);
				const data = await getTaskStatusHistory(taskId!);
				setStatusHistory(data);
			} catch (error) {
				notification.error({
					message: 'Error',
					description: (error as Error).message,
				});
			} finally {
				setLoading(false);
			}
		};

		fetchTaskStatusHistory();
	}, [taskId]);

	const columns = [
		{ title: 'ID', dataIndex: 'id', key: 'id' },
		{ title: 'Task ID', dataIndex: 'task_id', key: 'task_id' },
		{ title: 'Status', dataIndex: 'status', key: 'status' },
		{ title: 'Timestamp', dataIndex: 'timestamp', key: 'timestamp' },
		{ title: 'Duration', dataIndex: 'duration', key: 'duration' },
	];

	return (
		<div>
			<h2>Task Status History</h2>
			<Table dataSource={statusHistory} columns={columns} loading={loading} rowKey='id' />
		</div>
	);
};

export default TaskStatusHistoryPage;
