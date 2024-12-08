import React from 'react';
import { List, Card, Button } from 'antd';
import { useQuery } from '@tanstack/react-query';
import { useNavigate } from 'react-router-dom';
import { getAllProjects, getAllProjectsWithManagerId, getAllProjectsWithUserId } from '../services/projectService';
import { DTOCreateProject } from '../entities/models/CreateProject';
import { getTokenData, isManager, isMember } from '../utils/authHelpers';

const ProjectList: React.FC = () => {

  const navigate = useNavigate();

  const userData = getTokenData();

  const fetchProjects = async () => {
    if (isManager()) {
      return getAllProjectsWithManagerId(userData.id);
    } else if (isMember()) {
      return getAllProjectsWithUserId(userData.id);
    } else {
      throw new Error("User role is not recognized");
    }
  };

  const { data: projects, isLoading, error} = useQuery<DTOCreateProject[]>({
    queryKey: ['projects', userData.id],
    queryFn: fetchProjects,
    enabled: !!userData,
  });

  if (isLoading){
    return <p>Loading projects...</p>;
  }

  if (error){
    return <p>Error loading projects: {(error as Error).message}</p>;
  }

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
            <Button
              type="primary"
              onClick={() => navigate(`/project/${project.id}`)} 
            >
              Details
            </Button>
          </Card>
        </List.Item>
      )}
    />
  );
};

export default ProjectList;
