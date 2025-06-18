import { createFileRoute } from "@tanstack/react-router";
import { changelog } from "../changelog";
export const Route = createFileRoute("/changelog")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="min-h-screen bg-gray-50 py-8 px-4 sm:px-6 lg:px-8">
      <div className="max-w-3xl mx-auto">
        <h1 className="text-3xl font-bold text-gray-900 mb-8">Changelog</h1>

        {changelog.map((release) => (
          <div
            key={release.version}
            className="mb-8 bg-white p-6 rounded-lg shadow"
          >
            <div className="flex items-center justify-between mb-2">
              <h2 className="text-xl font-semibold text-gray-800">
                Version {release.version}
              </h2>
              <span className="text-sm text-gray-500">{release.date}</span>
            </div>
            <div className="mb-4 text-gray-500">{release.description}</div>
            <ul className="space-y-2">
              {release.changes.map((change, index) => (
                <li key={index} className="flex items-start">
                  <span className="text-red-500 mr-2">•</span>
                  <span className="text-gray-700">{change}</span>
                </li>
              ))}
            </ul>
          </div>
        ))}
      </div>
    </div>
  );
}
