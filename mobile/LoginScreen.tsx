import { useEffect, useState } from 'react';
import { StyleSheet, Text, View } from 'react-native';
import * as Google from 'expo-auth-session/providers/google';
import * as WebBrowser from 'expo-web-browser';
import { colors, spacing, typography } from './theme';
import { PrimaryButton } from './ui';

WebBrowser.maybeCompleteAuthSession();

// From the Google Cloud Console OAuth client created for ORBIT (Web
// application type - sufficient for verifying id_token server-side).
const GOOGLE_WEB_CLIENT_ID =
  '595683522917-al9clne9te05q8gb55a1t8v2t0kvf39u.apps.googleusercontent.com';

const BACKEND_AUTH_URL = 'http://localhost:8080/api/v1/auth/google';

type SignedInUser = {
  id: string;
  email: string;
  display_name: string;
};

type Props = {
  onSignedIn: (sessionToken: string, user: SignedInUser) => void;
};

export default function LoginScreen({ onSignedIn }: Props) {
  const [status, setStatus] = useState<string>('');

  const [request, response, promptAsync] = Google.useIdTokenAuthRequest({
    clientId: GOOGLE_WEB_CLIENT_ID,
  });

  useEffect(() => {
    if (response?.type !== 'success') {
      return;
    }

    const idToken = response.params.id_token;
    if (!idToken) {
      setStatus('Google did not return an id_token');
      return;
    }

    setStatus('Verifying with ORBIT backend...');

    fetch(BACKEND_AUTH_URL, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify({ id_token: idToken }),
    })
      .then(async (res) => {
        if (!res.ok) {
          const body = await res.json().catch(() => ({}));
          throw new Error(body.error ?? `backend returned ${res.status}`);
        }
        return res.json();
      })
      .then((data: { session_token: string; user: SignedInUser }) => {
        setStatus('Signed in');
        onSignedIn(data.session_token, data.user);
      })
      .catch((err) => {
        setStatus(`Sign-in failed: ${err.message}`);
      });
  }, [response]);

  return (
    <View style={styles.container}>
      <View style={styles.aura} />
      <View style={styles.brand}>
        <View style={styles.logoCircle}>
          <Text style={styles.logo}>🪐</Text>
        </View>
        <Text style={styles.title}>ORBIT</Text>
        <Text style={styles.tagline}>Spend freely, stay aware</Text>
      </View>
      <View style={styles.actions}>
        <PrimaryButton title="Sign in with Google" disabled={!request} onPress={() => promptAsync()} />
        {status ? <Text style={styles.status}>{status}</Text> : null}
      </View>
    </View>
  );
}

const styles = StyleSheet.create({
  container: {
    flex: 1,
    alignItems: 'center',
    justifyContent: 'center',
    backgroundColor: colors.surface,
    gap: spacing.xl,
    paddingHorizontal: spacing.margin,
  },
  aura: {
    position: 'absolute',
    top: -80,
    right: -80,
    width: 260,
    height: 260,
    borderRadius: 130,
    backgroundColor: colors.secondaryFixed,
    opacity: 0.18,
  },
  brand: { alignItems: 'center', gap: spacing.xs },
  logoCircle: {
    width: 72,
    height: 72,
    borderRadius: 36,
    backgroundColor: colors.primaryFixed,
    alignItems: 'center',
    justifyContent: 'center',
    marginBottom: spacing.xs,
  },
  logo: { fontSize: 34 },
  title: { ...typography.headlineXl, fontSize: 34, letterSpacing: 1, color: colors.onSurface },
  tagline: { ...typography.bodyMd, color: colors.onSurfaceVariant },
  actions: { width: '100%', maxWidth: 320, alignItems: 'center', gap: spacing.sm },
  status: { ...typography.bodySm, color: colors.onSurfaceVariant, textAlign: 'center' },
});
