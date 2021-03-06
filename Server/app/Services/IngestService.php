<?php

namespace App\Services;

use Log;
use Ramsey\Uuid\Uuid;
use Illuminate\Support\Facades\Cache;

use App\Models\User;
use App\Models\Card;
use App\Models\Deck;

use App\Services\ImageService;

class IngestService
{
    public function getDecks($userId): array
    {
        $decks = Deck::where("userId", $userId)->get();
        $output = [];
        foreach ($decks as $deck) {
            $output[] = [
                "Name" => $deck->name,
                "UID" => $deck->uid,
                "Commander" => $deck->commander,
                "Cards" => $deck->cards,
            ];
        }
        return $output;
    }

    public function countDecks($userId)
    {
        $count = Deck::where("userId", $userId)->count();
        return $count;
    }

    public function addCard(array $params, $front, $back)
    {
        $imageService = new ImageService();
        $slug = $params["slug"];
        $card = Card::where("slug", $slug)->first();
        if (empty($card)) {
            if (!is_null($front)) {
                $frontImageUid = $imageService->saveImage($front, 1);
            }
            if (!is_null($back)) {
                $backImageUid = $imageService->saveImage($back, 1);
            }
            $uid = Uuid::uuid4()->toString();
            $card = Card::create([
                "uid" => $uid,
                "name" => $params["name"],
                "slug" => $slug,
                "layout" => $params["layout"],
                "rarity" => $params["rarity"],
                "type" => $params["type"],
                "text" => $params["text"],
                "flavorText" => $params["flavorText"],
                "vitality" => $params["vitalitys"],
                "faceNames" => $params["faceNames"],
                "manaCosts" => $params["manaCosts"],
                "totalManaCosts" => $params["totalManaCosts"],
                "subtypes" => $params["subtypes"],
                "legalities" => $params["legalities"],
                "colors" => $params["colors"],
                "keywords" => $params["keywords"],
                "front" => $frontImageUid ?? null,
                "back" => $backImageUid ?? null,
            ]);
        } else {
            if (!is_null($front)) {
                $card->front = $imageService->saveImage($front, 1, $card->front);
            }
            if (!is_null($back)) {
                $card->back = $imageService->saveImage($back, 1, $card->back);
            }
            $card->name = $params["name"];
            $card->slug = $slug;
            $card->layout = $params["layout"];
            $card->rarity = $params["rarity"];
            $card->type = $params["type"];
            $card->text = $params["text"];
            $card->flavorText = $params["flavorText"];
            $card->vitality = $params["vitalitys"];
            $card->faceNames = $params["faceNames"];
            $card->manaCosts = $params["manaCosts"];
            $card->totalManaCosts = $params["totalManaCost"];
            $card->subtypes = $params["subtypes"];
            $card->legalities = $params["legalities"];
            $card->colors = $params["colors"];
            $card->keywords = $params["keywords"];
            $card->save();
        }
    }

    public function getAllUsers(): array
    {
        $output = [];
        $users = User::get();
        foreach ($users as $user) {
            $output[] = [
                "Name" => $user->name,
                "Email" => $user->email,
                "Uid" => $user->uid,
                "Suspended" => boolval($user->suspended),
                "Verified" => boolval($user->verified),
                "Admin" => boolval($user->admin),
            ];
        }
        return $output;
    }

    public function countUsers(): int
    {
        $count = Cache::get("user-count", null);
        if (is_null($count)) {
            $count = User::count();
        }
        return (int) $count;
    }

    public function countCards(): int
    {
        $count = Cache::get("cards-count", null);
        if (is_null($count)) {
            $count = Card::count();
        }
        return (int) $count;
    }
}
