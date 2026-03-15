import http from "k6/http";
import { check, sleep } from "k6";
import exec from "k6/execution";

const BASE_URL = (__ENV.BASE_URL || "http://localhost:8080/api/v1").replace(/\/+$/, "");
const VUS = Number.parseInt(__ENV.VUS || "10", 10);
const DURATION = __ENV.DURATION || "60s";
const SLEEP_SECONDS = Number.parseFloat(__ENV.SLEEP_SECONDS || "0.2");
const MAX_ERROR_RATE = __ENV.MAX_ERROR_RATE || "0.05";
const P95_MS = __ENV.P95_MS || "2000";

export const options = {
  vus: Number.isFinite(VUS) && VUS > 0 ? VUS : 10,
  duration: DURATION,
  summaryTrendStats: ["min", "avg", "med", "p(90)", "p(95)", "p(99)", "max"],
  thresholds: {
    http_req_failed: [`rate<${MAX_ERROR_RATE}`],
    http_req_duration: [`p(95)<${P95_MS}`],
  },
};

function asDateParam(date) {
  const year = date.getUTCFullYear();
  const month = String(date.getUTCMonth() + 1).padStart(2, "0");
  const day = String(date.getUTCDate()).padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function pickRoomCode(data) {
  if (Array.isArray(data.roomCodes) && data.roomCodes.length > 0) {
    return data.roomCodes[Math.floor(Math.random() * data.roomCodes.length)];
  }
  return data.roomCode;
}

export function setup() {
  const roomsURL = `${BASE_URL}/rooms`;
  const roomsResponse = http.get(roomsURL, {
    tags: { endpoint: "rooms_list", phase: "setup" },
  });

  const roomCodeFromEnv = (__ENV.ROOM_CODE || "").trim();
  const isRoomsOK = check(roomsResponse, {
    "setup rooms status is 200": (r) => r.status === 200,
  });

  if (!isRoomsOK) {
    throw new Error(`setup failed: GET ${roomsURL} returned ${roomsResponse.status}`);
  }

  const payload = roomsResponse.json();
  const roomCodes = [];
  if (payload && Array.isArray(payload.rooms)) {
    for (const room of payload.rooms) {
      if (room && typeof room.code === "string" && room.code.trim() !== "") {
        roomCodes.push(room.code.trim());
      }
    }
  }

  const roomCode = roomCodeFromEnv || roomCodes[0] || "";

  const now = new Date();
  const startDate = asDateParam(now);
  const endDate = asDateParam(now);

  return {
    roomCode,
    roomCodes,
    hasRoomCode: roomCode !== "",
    startDate,
    endDate,
  };
}

export default function (data) {
  const roomCode = pickRoomCode(data);
  const hasRoomCode = data && data.hasRoomCode === true;
  const encodedCode = encodeURIComponent(roomCode);

  const roll = Math.random();
  if (!hasRoomCode || roll < 0.55) {
    const response = http.get(`${BASE_URL}/rooms`, {
      tags: { endpoint: "rooms_list", scenario: exec.scenario.name },
    });
    check(response, {
      "rooms list status is 200": (r) => r.status === 200,
    });
  } else if (roll < 0.85) {
    const response = http.get(`${BASE_URL}/rooms/${encodedCode}`, {
      tags: { endpoint: "room_detail", scenario: exec.scenario.name },
    });
    check(response, {
      "room detail status is 200": (r) => r.status === 200,
    });
  } else {
    const response = http.get(
      `${BASE_URL}/rooms/${encodedCode}/schedule?start=${data.startDate}&end=${data.endDate}`,
      { tags: { endpoint: "room_schedule", scenario: exec.scenario.name } },
    );
    check(response, {
      "room schedule status is 200": (r) => r.status === 200,
    });
  }

  if (Number.isFinite(SLEEP_SECONDS) && SLEEP_SECONDS > 0) {
    sleep(SLEEP_SECONDS);
  }
}
