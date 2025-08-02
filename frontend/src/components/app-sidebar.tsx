import { Search, Plus, Folder, MessageCircle } from "lucide-react";
import { useNavigate } from "react-router-dom";
import { useState, useEffect } from "react";
import { useStore } from "@nanostores/react";
import {
  Dialog,
  DialogContent,
  DialogFooter,
  DialogHeader,
  DialogTitle,
  DialogTrigger,
} from "@/components/ui/dialog";
import { Button } from "@/components/ui/button";
import { Input } from "@/components/ui/input";
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
import {
  $chatList,
  $currentChatId,
  $currentProject,
  $currentProjectId,
  $projectList,
  $searchText,
  $searchResults,
  createNewChat,
  createProject,
  getProjectList,
} from "@/store/chat";

export function AppSidebar() {
  const projectsList = useStore($projectList);
  const chatsList = useStore($chatList);
  const searchResults = useStore($searchResults);
  
  const [projectName, setProjectName] = useState("");
  const [isProjectDialogOpen, setIsProjectDialogOpen] = useState(false);
  const [isSearchDialogOpen, setIsSearchDialogOpen] = useState(false);
  const [localSearchText, setLocalSearchText] = useState("");

  const navigate = useNavigate();

  // Debounced effect to update $searchText when user stops typing
  useEffect(() => {
    const timeoutId = setTimeout(() => {
      if (localSearchText.trim()) {
        $searchText.set(localSearchText.trim());
      }
    }, 500); // 500ms delay after user stops typing

    return () => clearTimeout(timeoutId);
  }, [localSearchText]);

  // Reset local search state when modal opens
  useEffect(() => {
    if (isSearchDialogOpen) {
      setLocalSearchText("");
    }
  }, [isSearchDialogOpen]);

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
    navigate(`/project/${projectId}`, { replace: true });
  };

  const handleChatSelect = (selectedChatId: string) => {
    $currentChatId.set(selectedChatId);
    navigate(`/chat/${selectedChatId}`, { replace: true });
  };

  const handleCreateProject = async () => {
    await createProject(projectName, "description");
    await getProjectList();
    setProjectName("");
    setIsProjectDialogOpen(false);
  };

  const handleCancelProject = () => {
    setProjectName("");
    setIsProjectDialogOpen(false);
  };

  const handleSearchClose = () => {
    setIsSearchDialogOpen(false);
    $searchText.set("");
    setLocalSearchText("");
  };

  const handleSearchResultClick = (chatId: string) => {
    navigate(`/chat/${chatId}`);
    handleSearchClose();
  };

  return (
    <Sidebar className="pl-4">
      <SidebarContent className="overflow-y-auto overflow-x-hidden h-full">
        <SidebarGroup>
          <SidebarGroupLabel className="w-full h-[25%] text-2xl">
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
                <Dialog open={isSearchDialogOpen} onOpenChange={setIsSearchDialogOpen}>
                  <DialogTrigger asChild>
                    <SidebarMenuButton>
                      <Search />
                      <span>Search Chats</span>
                    </SidebarMenuButton>
                  </DialogTrigger>
                  <DialogContent className="max-w-2xl max-h-[80vh] p-0">
                    <DialogHeader className="px-6 pt-6 pb-4">
                      <DialogTitle>Search Conversations</DialogTitle>
                    </DialogHeader>
                    
                    <div className="px-6">
                      <Input
                        type="text"
                        placeholder="Search conversations..."
                        value={localSearchText}
                        onChange={(e) => setLocalSearchText(e.target.value)}
                        autoFocus
                        className="w-full"
                      />
                    </div>

                    {/* Search Results */}
                    <div className="flex-1 overflow-y-auto px-6 pb-6 mt-4 space-y-2 max-h-96">
                      {searchResults.length > 0 ? (
                        searchResults.map((result, index) => (
                          <div
                            key={index}
                            className="p-3 border border-gray-200 dark:border-gray-600 rounded-lg hover:bg-gray-50 dark:hover:bg-gray-700 cursor-pointer transition-colors"
                            onClick={() => handleSearchResultClick(result.chat_id)}
                          >
                            <div className="text-sm text-gray-600 dark:text-gray-400 mb-1">
                              Chat: {result.chat_name || "Unnamed Chat"}
                            </div>
                            <div className="text-gray-900 dark:text-white line-clamp-2">
                              {result.matched_text}
                            </div>
                          </div>
                        ))
                      ) : localSearchText.trim() ? (
                        <div className="text-center text-gray-500 dark:text-gray-400 py-8">
                          No results found for "{localSearchText}"
                        </div>
                      ) : (
                        <div className="text-center text-gray-500 dark:text-gray-400 py-8">
                          Start typing to search your conversations...
                        </div>
                      )}
                    </div>
                  </DialogContent>
                </Dialog>
              </SidebarMenuItem>
              
              <SidebarMenuItem>
                <Dialog open={isProjectDialogOpen} onOpenChange={setIsProjectDialogOpen}>
                  <DialogTrigger asChild>
                    <SidebarMenuButton>
                      <Plus />
                      <span>Create Project</span>
                    </SidebarMenuButton>
                  </DialogTrigger>
                  <DialogContent>
                    <DialogHeader>
                      <DialogTitle>Create New Project</DialogTitle>
                    </DialogHeader>
                    <div className="grid gap-4 py-4">
                      <div className="grid grid-cols-3 items-center gap-4">
                        <Input
                          id="project-name"
                          value={projectName}
                          onChange={(e) => setProjectName(e.target.value)}
                          placeholder="Enter project name"
                          className="col-span-4"
                        />
                      </div>
                    </div>
                    <DialogFooter>
                      <Button
                        type="button"
                        variant="outline"
                        onClick={handleCancelProject}
                      >
                        Cancel
                      </Button>
                      <Button
                        type="button"
                        onClick={handleCreateProject}
                        disabled={!projectName.trim()}
                      >
                        Create
                      </Button>
                    </DialogFooter>
                  </DialogContent>
                </Dialog>
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
                <SidebarMenuItem key={chat.chatId}>
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