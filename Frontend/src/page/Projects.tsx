import React from 'react';
import { List, Card, Button} from 'antd';
import { getAllProjects } from '../services/projectService';
import { useQuery } from '@tanstack/react-query';
import { DTOCreateProject } from '../entities/models/CreateProject';

const ProjectList: React.FC = () => {
  const { data: projects} = useQuery<DTOCreateProject[]>({
		queryKey: ['projects'],
		queryFn: () => getAllProjects(),
	});


  
if (!projects || projects.length === 0) {
	return <p>No projects available.</p>; 
}

  return (
    <List
      grid={{ gutter: 16, column: 4 }}
      dataSource={projects}
      renderItem={(project) => (
        <List.Item>
          <Card title={project.name}>
            <p><strong>Manager:</strong> {project.manager?.username}</p>
            <p><strong>Members:</strong> {project.members?.map(member => member.username).join(', ')}</p>
            <p><strong>Expected End Date:</strong> {project.expectedEndDate}</p>
            <p><strong>Min Members:</strong> {project.minMembers}</p>
            <p><strong>Max Members:</strong> {project.maxMembers}</p>
            <Button type="primary">Details</Button>
          </Card>
        </List.Item>
      )}
    />
  );
};

export default ProjectList;
