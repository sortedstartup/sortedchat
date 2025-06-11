import { type RouteConfig, index, route } from "@react-router/dev/routes";

export default [
  index("routes/home.tsx"),
  route("chat/:id", "routes/chat.$id.tsx"),
  route("upload", "routes/upload.tsx")
] satisfies RouteConfig;
