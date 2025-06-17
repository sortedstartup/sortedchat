import { type RouteConfig, index, route } from "@react-router/dev/routes";

export default [
  index("routes/home.tsx"),
  route("chat/:id", "routes/chat.$id.tsx"),
  route("project/:id", "routes/project.tsx")
] satisfies RouteConfig;
