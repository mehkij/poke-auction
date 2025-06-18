import { createRootRoute, Link, Outlet } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";

export const Route = createRootRoute({
  component: () => (
    <>
      <nav className="p-4 flex gap-4">
        <Link
          to="/"
          className="[&.active]:bg-gray-400/25 rounded p-2 transition-all"
        >
          Home
        </Link>{" "}
        <Link
          to="/changelog"
          className="[&.active]:bg-gray-400/25 rounded p-2 transition-all"
        >
          Changelog
        </Link>
      </nav>
      <Outlet />
      <TanStackRouterDevtools />
    </>
  ),
});
