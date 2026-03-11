const net = require("net");
const path = require("path");
const ipcModule = require("@node-ipc/node-ipc");

const ipc = ipcModule.default || ipcModule;

function createLogger(levelName = "info") {
  const levels = new Map([
    ["error", 0],
    ["info", 1],
    ["debug", 2],
  ]);
  const configuredLevel = levels.get(String(levelName).toLowerCase()) ?? levels.get("info");

  function logAt(name, threshold, message, meta) {
    if (configuredLevel < threshold) {
      return;
    }

    const prefix = `[leapp-shim] ${name.toUpperCase()}`;
    if (meta === undefined) {
      console.error(`${prefix} ${message}`);
      return;
    }
    console.error(`${prefix} ${message}`, meta);
  }

  return {
    error(message, meta) {
      logAt("error", levels.get("error"), message, meta);
    },
    info(message, meta) {
      logAt("info", levels.get("info"), message, meta);
    },
    debug(message, meta) {
      logAt("debug", levels.get("debug"), message, meta);
    },
  };
}

function getConfig(env = process.env) {
  const tcpPort = Number.parseInt(env.LEAPP_SHIM_TCP_PORT ?? "43827", 10);
  if (!Number.isInteger(tcpPort) || tcpPort < 1 || tcpPort > 65535) {
    throw new Error(`invalid LEAPP_SHIM_TCP_PORT: ${env.LEAPP_SHIM_TCP_PORT}`);
  }

  return {
    serverId: env.LEAPP_SHIM_SERVER_ID || "leapp_da",
    tcpHost: env.LEAPP_SHIM_TCP_HOST || "127.0.0.1",
    tcpPort,
    logLevel: env.LEAPP_SHIM_LOG_LEVEL || "info",
    socketRoot: normalizeSocketRoot(env.LEAPP_SHIM_SOCKET_ROOT),
    appspace: env.LEAPP_SHIM_APPSPACE,
  };
}

function normalizeSocketRoot(socketRoot) {
  if (!socketRoot) {
    return socketRoot;
  }
  return socketRoot.endsWith(path.sep) ? socketRoot : `${socketRoot}${path.sep}`;
}

function closeSocket(socket) {
  if (!socket || socket.destroyed) {
    return;
  }
  socket.destroy();
}

function normalizeBuffer(data) {
  return Buffer.isBuffer(data) ? data : Buffer.from(data);
}

function createShim(options = {}) {
  const config = options.config || getConfig();
  const logger = options.logger || createLogger(config.logLevel);
  const ipcInstance = options.ipc || ipc;
  const netModule = options.net || net;
  const activeSockets = new Set();

  ipcInstance.config.id = config.serverId;
  ipcInstance.config.silent = true;
  ipcInstance.config.rawBuffer = true;
  ipcInstance.config.retry = 0;

  if (config.socketRoot) {
    ipcInstance.config.socketRoot = config.socketRoot;
  }
  if (config.appspace !== undefined) {
    ipcInstance.config.appspace = config.appspace;
  }

  function registerSocket(socket) {
    activeSockets.add(socket);
    socket.once("close", () => {
      activeSockets.delete(socket);
    });
  }

  function bridgeConnection(clientSocket) {
    const upstream = netModule.createConnection({
      host: config.tcpHost,
      port: config.tcpPort,
    });
    registerSocket(upstream);
    registerSocket(clientSocket);

    let closing = false;
    function closePair() {
      if (closing) {
        return;
      }
      closing = true;
      closeSocket(clientSocket);
      closeSocket(upstream);
    }

    upstream.on("connect", () => {
      logger.info(`connected ${config.serverId} client to tcp://${config.tcpHost}:${config.tcpPort}`);
    });

    upstream.on("data", (chunk) => {
      ipcInstance.server.emit(clientSocket, normalizeBuffer(chunk));
    });

    upstream.on("error", (error) => {
      logger.error("upstream tcp connection failed", error.message);
      closePair();
    });

    upstream.on("close", () => {
      logger.debug("upstream tcp connection closed");
      closePair();
    });

    clientSocket.on("data", (chunk) => {
      if (!upstream.destroyed) {
        upstream.write(normalizeBuffer(chunk));
      }
    });

    clientSocket.on("error", (error) => {
      logger.error("node-ipc client socket failed", error.message);
      closePair();
    });

    clientSocket.on("close", () => {
      logger.debug("node-ipc client socket closed");
      closePair();
    });
  }

  return {
    config,
    logger,
    async start() {
      await new Promise((resolve) => {
        ipcInstance.serve(() => {
          ipcInstance.server.on("connect", bridgeConnection);
          logger.info(`node-ipc shim listening as ${config.serverId}`);
          resolve();
        });
        ipcInstance.server.start();
      });
    },
    async stop() {
      for (const socket of Array.from(activeSockets)) {
        closeSocket(socket);
      }

      await new Promise((resolve) => {
        ipcInstance.server.off("connect", bridgeConnection);
        ipcInstance.server.stop();
        setImmediate(resolve);
      });
    },
  };
}

async function main() {
  const shim = createShim();
  await shim.start();

  const stop = async () => {
    await shim.stop();
    process.exit(0);
  };

  process.once("SIGINT", stop);
  process.once("SIGTERM", stop);
}

if (require.main === module) {
  main().catch((error) => {
    console.error("[leapp-shim] ERROR failed to start", error);
    process.exit(1);
  });
}

module.exports = {
  createLogger,
  createShim,
  getConfig,
};
