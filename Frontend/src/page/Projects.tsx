import React, { useEffect, useState } from 'react';
import { List, Card, Button, notification } from 'antd';
import { DTOCreateProject } from '../entities/models/CreateProject'; 
import { getAllProjects } from '../services/projectService';

const ProjectList: React.FC = () => {
  const [projects, setProjects] = useState<DTOCreateProject[]>([]);

  useEffect(() => {
    const fetchProjects = async () => {
      try {
        const response = await getAllProjects();
        
        if (response && response.data) {
          setProjects(response.data); 
        } else {
          notification.error({
            message: 'Error',
            description: 'No project data available.',
          });
        }
      } catch (error) {
        notification.error({
          message: 'Error',
          description: 'Failed to load projects.',
        });
      }
    };
  
    fetchProjects();
  }, []);
  

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
            <Button type="primary">Add Member</Button>
          </Card>
        </List.Item>
      )}
    />
  );
};

export default ProjectList;
