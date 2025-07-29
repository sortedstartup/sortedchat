import { Search, Plus, FolderPlus, Folder, MessageCircle } from "lucide-react";
import { useNavigate, useLocation } from "react-router-dom";

import {
  Sidebar,
  SidebarContent,
  SidebarGroup,
  SidebarGroupContent,
  SidebarGroupLabel,
  SidebarMenu,
  SidebarMenuButton,
  SidebarMenuItem,
  SidebarSeparator,
} from "@/components/ui/sidebar";
import { $chatList, $currentChatId, $currentProject, $currentProjectId, $projectList, createNewChat } from "@/store/chat";
import { useStore } from "@nanostores/react";

export function AppSidebar() {
  const projectsList = useStore($projectList);
  const chatsList = useStore($chatList);

  let navigate: ReturnType<typeof useNavigate> | undefined;
  let location: ReturnType<typeof useLocation> | undefined;
  try {
    navigate = useNavigate();
    location = useLocation();
  } catch {
    navigate = undefined;
    location = undefined;
  }

  const handleNewChat = async () => {
    try {
      const chatId = await createNewChat();
      if (chatId && navigate) {
        navigate(`/chat/${chatId}`);
      } else if (!navigate) {
        window.location.href = `/chat/${chatId}`;
      } else {
        console.error("No chatId returned from server");
      }
    } catch (err) {
      console.error("Failed to create new chat:", err);
    }
  };

  const handleProjectClick = (projectId: string) => {
    $currentProject.set(
      projectsList.find((p) => p.id === projectId)?.name || ""
    );
    $currentProjectId.set(projectId);
    if (navigate && location) {
      if (location.pathname !== `/project/${projectId}`) {
        navigate(`/project/${projectId}`, { replace: true });
      }
    } else {
      window.location.href = `/project/${projectId}`;
    }
  };

  const handleChatSelect = (selectedChatId: string) => {
    $currentChatId.set(selectedChatId);
    if (navigate && location) {
      if (location.pathname !== `/chat/${selectedChatId}`) {
        navigate(`/chat/${selectedChatId}`, { replace: true });
      }
    } else {
      window.location.href = `/chat/${selectedChatId}`;
    }
  };

  return (
    <Sidebar className="pl-4">
      <SidebarContent className="overflow-y-auto overflow-x-hidden h-full">
        <SidebarGroup>
          <SidebarGroupLabel
            className="w-full h-[25%] text-2xl
"
          >
            SortedChat
          </SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              <SidebarMenuItem>
                <SidebarMenuButton asChild>
                  <button onClick={handleNewChat}>
                    <Plus />
                    <span>New Chat</span>
                  </button>
                </SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem>
                <SidebarMenuButton asChild>
                  <a href="#">
                    <Search />
                    <span>Search Chats</span>
                  </a>
                </SidebarMenuButton>
              </SidebarMenuItem>
              <SidebarMenuItem>
                <SidebarMenuButton asChild>
                  <a href="#">
                    <FolderPlus />
                    <span>Create Project</span>
                  </a>
                </SidebarMenuButton>
              </SidebarMenuItem>
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        <SidebarSeparator />

        <SidebarGroup>
          <SidebarGroupLabel className="text-xs font-semibold text-muted-foreground mb-1">
            Projects
          </SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {projectsList.map((project) => (
                <SidebarMenuItem key={project.name}>
                  <SidebarMenuButton
                    onClick={() => handleProjectClick(project.id)}
                  >
                    <Folder />
                    <span>{project.name}</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>

        <SidebarSeparator />

        <SidebarGroup>
          <SidebarGroupLabel className="text-xs font-semibold text-muted-foreground mb-1">
            Chats
          </SidebarGroupLabel>
          <SidebarGroupContent>
            <SidebarMenu>
              {chatsList.map((chat) => (
                <SidebarMenuItem key={chat.name}>
                  <SidebarMenuButton
                    onClick={() => handleChatSelect(chat.chatId)}
                  >
                    <MessageCircle />
                    <span>{chat.name}</span>
                  </SidebarMenuButton>
                </SidebarMenuItem>
              ))}
            </SidebarMenu>
          </SidebarGroupContent>
        </SidebarGroup>
      </SidebarContent>
    </Sidebar>
  );
}
