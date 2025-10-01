import { createRootRoute, Link, Outlet } from "@tanstack/react-router";
import { TanStackRouterDevtools } from "@tanstack/react-router-devtools";
import ThemeToggle from "../components/ThemeToggle";

export const Route = createRootRoute({
  component: RootComponent,
});

function RootComponent() {
  return (
    <>
      <nav className="p-4 flex justify-between items-center">
        <div className="flex gap-4">
          <Link
            to="/"
            className="[&.active]:bg-gray-400/25 rounded p-2 transition-all"
            viewTransition
          >
            Home
          </Link>{" "}
          <Link
            to="/changelog"
            className="[&.active]:bg-gray-400/25 rounded p-2 transition-all"
            viewTransition
          >
            Changelog
          </Link>
          <Link
            to="/updates"
            className="[&.active]:bg-gray-400/25 rounded p-2 transition-all"
            viewTransition
          >
            Updates
          </Link>
        </div>
        <ThemeToggle />
      </nav>

      <div className="transition-opacity">
        <Outlet />
      </div>
      <TanStackRouterDevtools />
    </>
  );
}
