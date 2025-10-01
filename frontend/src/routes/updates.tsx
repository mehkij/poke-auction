import { createFileRoute } from "@tanstack/react-router";
import { upcoming } from "../upcoming";
export const Route = createFileRoute("/updates")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="min-h-screen py-8 px-4 sm:px-6 lg:px-8 transition-colors">
      <div className="max-w-3xl mx-auto">
        <h1 className="text-3xl font-bold dark:text-white text-gray-900 mb-8">
          Upcoming Features & Bugfixes
        </h1>

        <ul className="space-y-4">
          {upcoming.map((update, index) => (
            <li key={index} className="flex items-start">
              <span className="text-red-500 mr-2">•</span>
              <span className="dark:text-white text-gray-700">
                {update.featureDesc}
              </span>
            </li>
          ))}
        </ul>
      </div>
    </div>
  );
}
