import { useQuery } from "@tanstack/react-query";

async function fetchBotHealth() {
  const res = await fetch("http://18.225.92.36:8080/api/status");
  if (!res.ok) {
    throw new Error("Network response was not ok");
  }
  return res.json();
}

function Status() {
  const { status, data, error } = useQuery({
    queryKey: ["bot-status"],
    queryFn: fetchBotHealth,
    refetchInterval: 300000,
  });

  if (status === "error") {
    return <span>Error fetching bot status: {error.message}</span>;
  }

  const online = data?.status === "online";

  return (
    <div className="flex items-center m-4 gap-2">
      <div
        className={`w-2 h-2 rounded-full ${
          online ? "bg-green-500" : "bg-red-500"
        } transition-colors`}
      ></div>
      <p>
        The PokéAuction bot is currently{" "}
        {online ? (
          <span className="text-green-500 transition-colors">online</span>
        ) : (
          <span className="text-red-500 transition-colors">offline</span>
        )}
        !
      </p>
    </div>
  );
}

export default Status;
