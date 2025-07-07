import { createFileRoute } from "@tanstack/react-router";
import Button from "../components/Button";
import Status from "../components/Status";
import GHLogo from "../assets/github-mark.svg";
import DiscordLogo from "../assets/Discord-Symbol-White.svg";
import PokeBall from "../assets/Poke_Ball.webp";

export const Route = createFileRoute("/")({
  component: RouteComponent,
});

function RouteComponent() {
  return (
    <div className="h-screen w-screen flex flex-col justify-center items-center">
      <img className="size-64 m-4" src={PokeBall}></img>
      <Status />
      <div className="flex m-4 gap-4">
        <Button
          className="flex gap-2 items-center justify-center bg-red-500 hover:bg-red-600 text-white border-2 border-red-600"
          onClick={() =>
            window.open(
              "https://discord.com/oauth2/authorize?client_id=1363982474270736414&permissions=277025516544&integration_type=0&scope=applications.commands+bot",
              "_blank"
            )
          }
        >
          <img className="size-4" src={DiscordLogo}></img>
          <p>Invite</p>
        </Button>

        <Button
          className="flex gap-2 items-center justify-center text-black bg-white hover:bg-gray-100 border-2 border-gray-100"
          onClick={() =>
            window.open("https://github.com/mehkij/poke-auction", "_blank")
          }
        >
          <img className="size-4" src={GHLogo}></img>
          <p>GitHub</p>
        </Button>
      </div>
      <div className="mt-4">
        <p className="text-red-500 font-bold">
          ATTENTION: The bot has temporarily been reverted to Version 1.1.3
          while an issue with the /config command is being fixed. This means
          /config will be unavailable for use at this time.
        </p>
      </div>
    </div>
  );
}
