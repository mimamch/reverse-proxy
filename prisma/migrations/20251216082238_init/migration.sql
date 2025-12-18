-- CreateTable
CREATE TABLE "proxies" (
    "id" TEXT NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,

    CONSTRAINT "proxies_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "hosts" (
    "id" TEXT NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "host" TEXT NOT NULL,
    "proxy_id" TEXT,

    CONSTRAINT "hosts_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "backends" (
    "id" TEXT NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "scheme" TEXT NOT NULL,
    "host" TEXT NOT NULL,
    "port" INTEGER,
    "proxy_id" TEXT,

    CONSTRAINT "backends_pkey" PRIMARY KEY ("id")
);

-- CreateTable
CREATE TABLE "headers" (
    "id" TEXT NOT NULL,
    "created_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "updated_at" TIMESTAMP(3) NOT NULL DEFAULT CURRENT_TIMESTAMP,
    "key" TEXT NOT NULL,
    "value" TEXT NOT NULL,
    "proxy_id" TEXT,

    CONSTRAINT "headers_pkey" PRIMARY KEY ("id")
);

-- AddForeignKey
ALTER TABLE "hosts" ADD CONSTRAINT "hosts_proxy_id_fkey" FOREIGN KEY ("proxy_id") REFERENCES "proxies"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "backends" ADD CONSTRAINT "backends_proxy_id_fkey" FOREIGN KEY ("proxy_id") REFERENCES "proxies"("id") ON DELETE CASCADE ON UPDATE CASCADE;

-- AddForeignKey
ALTER TABLE "headers" ADD CONSTRAINT "headers_proxy_id_fkey" FOREIGN KEY ("proxy_id") REFERENCES "proxies"("id") ON DELETE CASCADE ON UPDATE CASCADE;
