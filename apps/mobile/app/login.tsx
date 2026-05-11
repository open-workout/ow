import { View, Text, TouchableOpacity, StatusBar, KeyboardAvoidingView, Platform, ScrollView } from 'react-native';
import { SafeAreaView } from 'react-native-safe-area-context';
import { LinearGradient } from 'expo-linear-gradient';
import { Ionicons } from '@expo/vector-icons';
import { useRouter } from 'expo-router';
import { useState } from 'react';
import FormField from './components/FormField';

export default function LoginScreen() {
  const router = useRouter();
  const [identifier, setIdentifier] = useState('');
  const [password, setPassword] = useState('');
  const [showPassword, setShowPassword] = useState(false);

  return (
    <SafeAreaView style={{ flex: 1, backgroundColor: '#0a0a0a' }} edges={['top']}>
      <StatusBar barStyle="light-content" backgroundColor="#0a0a0a" />
      <KeyboardAvoidingView behavior={Platform.OS === 'ios' ? 'padding' : 'height'} style={{ flex: 1 }}>
        <ScrollView
          contentContainerStyle={{ flexGrow: 1 }}
          keyboardShouldPersistTaps="handled"
          showsVerticalScrollIndicator={false}
        >
          <View style={{ paddingHorizontal: 24, paddingTop: 8 }}>
            <TouchableOpacity
              onPress={() => router.back()}
              style={{ width: 40, height: 40, borderRadius: 20, backgroundColor: '#18181b', alignItems: 'center', justifyContent: 'center', borderWidth: 1, borderColor: '#27272a' }}
            >
              <Ionicons name="arrow-back" size={20} color="#a1a1aa" />
            </TouchableOpacity>
          </View>

          <View style={{ paddingHorizontal: 24, paddingTop: 40, flex: 1 }}>
            <Text style={{ color: '#fff', fontSize: 28, fontWeight: '800', letterSpacing: -0.5, marginBottom: 8 }}>
              Welcome back
            </Text>
            <Text style={{ color: '#71717a', fontSize: 16, marginBottom: 40, lineHeight: 24 }}>
              Sign in to continue tracking your progress.
            </Text>

            <FormField
              label="Username or Email"
              value={identifier}
              onChangeText={setIdentifier}
              placeholder="Enter your username or email"
              autoCapitalize="none"
              keyboardType="email-address"
              autoCorrect={false}
            />

            <FormField
              label="Password"
              value={password}
              onChangeText={setPassword}
              placeholder="Enter your password"
              secureTextEntry={!showPassword}
              rightIcon={
                <TouchableOpacity onPress={() => setShowPassword((v) => !v)}>
                  <Ionicons name={showPassword ? 'eye-off-outline' : 'eye-outline'} size={20} color="#71717a" />
                </TouchableOpacity>
              }
            />

            <TouchableOpacity style={{ alignSelf: 'flex-end', marginTop: -10, marginBottom: 36 }}>
              <Text style={{ color: '#71717a', fontSize: 14 }}>Forgot password?</Text>
            </TouchableOpacity>

            <TouchableOpacity activeOpacity={0.85} onPress={() => router.replace('/home')}>
              <LinearGradient
                colors={['#f4f4f5', '#a1a1aa']}
                start={{ x: 0, y: 0 }}
                end={{ x: 1, y: 1 }}
                style={{ borderRadius: 16, paddingVertical: 17, alignItems: 'center' }}
              >
                <Text style={{ color: '#09090b', fontWeight: '700', fontSize: 17 }}>Sign In</Text>
              </LinearGradient>
            </TouchableOpacity>

            <View style={{ flexDirection: 'row', justifyContent: 'center', marginTop: 32 }}>
              <Text style={{ color: '#71717a', fontSize: 14 }}>Don't have an account? </Text>
              <TouchableOpacity onPress={() => router.replace('/signup')}>
                <Text style={{ color: '#f4f4f5', fontSize: 14, fontWeight: '600' }}>Create one</Text>
              </TouchableOpacity>
            </View>
          </View>
        </ScrollView>
      </KeyboardAvoidingView>
    </SafeAreaView>
  );
}
