import { AlertTriangle, Loader2 } from "lucide-react";
import { Card, CardContent } from "@/components/ui/card";

export function LoadingState({ label = "Loading data" }: { label?: string }) {
  return (
    <Card>
      <CardContent className="flex items-center gap-3 p-6 text-sm text-muted-foreground">
        <Loader2 className="h-4 w-4 animate-spin" />
        {label}
      </CardContent>
    </Card>
  );
}

export function ErrorState({ message }: { message: string }) {
  return (
    <Card className="border-red-200 bg-red-50">
      <CardContent className="flex items-center gap-3 p-6 text-sm text-red-800">
        <AlertTriangle className="h-4 w-4" />
        {message}
      </CardContent>
    </Card>
  );
}
