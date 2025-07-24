import React from 'react';
import { List, Spin, Typography, notification } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { getActivityHistory } from '../services/activityService';

const { Title, Text } = Typography;

const ActivityHistory: React.FC = () => {
	const { data, isLoading, isError, error } = useQuery({
		queryKey: ['activity-history'],
		queryFn: getActivityHistory,
	});

	if (isLoading) {
		return <Spin tip="Loading activity history..." />;
	}

	if (isError) {
		notification.error({
			message: 'Error fetching activity history',
			description: (error as Error).message,
		});
		return null;
	}

	return (
		<div style={{ padding: '24px' }}>
			<Title level={3}>Activity History</Title>
			<List
				bordered
				dataSource={data}
				renderItem={(item: any) => {
					const description = item.details?.description || item.details?.fileName;

					return (
						<List.Item>
							<div>
								<Text strong>Activity Type:</Text> {item.activityType} <br />
								{description && (
									<>
										<Text strong>Description:</Text> {description} <br />
									</>
								)}
								{item.taskId && (
									<>
										<Text strong>Task ID:</Text> {item.taskId} <br />
									</>
								)}
								{item.projectId && (
									<>
										<Text strong>Project ID:</Text> {item.projectId} <br />
									</>
								)}
								{item.timestamp && (
									<>
										<Text strong>Time:</Text>{' '}
										{new Date(item.timestamp).toLocaleString()}
									</>
								)}
							</div>
						</List.Item>
					);
				}}
			/>
		</div>
	);
};

export default ActivityHistory;