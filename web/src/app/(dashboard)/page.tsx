"use client";

import { Card, Typography } from "antd";
import { useCurrentUser } from "@/lib/hooks/useCurrentUser";

const { Title, Paragraph } = Typography;

export default function DashboardHomePage() {
  const { user } = useCurrentUser();

  return (
    <Card>
      <Title level={3}>Welcome{user ? `, ${user.fullName}` : ""}</Title>
      <Paragraph type="secondary">
        This is Phase 1 of Digital Memorial — foundation, setup, and site settings are in place. Cemetery
        management modules (tombstones, columbarium, ancestor plaques, inscriptions) and user management arrive in
        later phases.
      </Paragraph>
    </Card>
  );
}
