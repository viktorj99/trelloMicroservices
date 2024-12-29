import React from 'react';
import ReactFlow, {
	Controls,
	Background,
	Edge,
	Node,
	Position,
	BackgroundVariant,
} from 'reactflow';
import 'reactflow/dist/style.css';
import { TaskWithDependencies } from '../../entities/models/TaskWithDependencies';

interface Props {
	data: TaskWithDependencies[];
}

const TaskGraph: React.FC<Props> = ({ data }) => {
	// Convert tasks to nodes and edges
	const nodes: Node[] = data.map((task, index) => ({
		id: task.id,
		data: { label: task.title },
		position: { x: index * 150, y: index * 100 },
		sourcePosition: Position.Right,
		targetPosition: Position.Left,
	}));

	const edges: Edge[] = [];

	data.forEach((task) => {
		if (task.dependencies) {
			task.dependencies.forEach((dependency) => {
				edges.push({
					id: `e${task.id}-${dependency.id}`,
					source: task.id,
					target: dependency.id!,
				});
			});
		}
	});

	return (
		<div style={{ height: '500px', width: '100%' }}>
			<ReactFlow nodes={nodes} edges={edges} fitView>
				<Controls />
				<Background variant={BackgroundVariant.Dots} gap={12} size={1} />
			</ReactFlow>
		</div>
	);
};

export default TaskGraph;
