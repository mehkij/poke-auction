type StatusProps = {
  online: boolean;
};

function Status({ online }: StatusProps) {
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
